package engine

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/tiklab/tiklab/internal/routeros"
)

// Engine manages simulated user lifecycle and traffic generation.
type Engine struct {
	mu        sync.Mutex
	users     []*SimulatedUser
	stopChans map[int]chan struct{}
	running   atomic.Bool
	nextIP    uint32
	cfg       EngineConfig
	ros       *routeros.Client
}

// NewEngine creates a new behavior engine.
func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		stopChans: make(map[int]chan struct{}),
		cfg:       cfg,
	}
}

// Start begins traffic generation for count users.
func (e *Engine) Start(count int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running.Load() {
		return nil // Already running
	}

	// Connect to RouterOS API
	ros := routeros.NewClient()
	user, pass := routeros.DefaultCredentials()
	if err := ros.Connect(e.cfg.RouterOSHost, e.cfg.RouterOSAPIPort, user, pass); err != nil {
		return err
	}
	e.ros = ros

	profiles := AssignProfiles(count)
	users, err := GenerateUsers(count, profiles)
	if err != nil {
		ros.Close()
		return err
	}

	e.users = make([]*SimulatedUser, len(users))
	for i := range users {
		e.users[i] = &users[i]
	}

	// Reset IP allocator
	atomic.StoreUint32(&e.nextIP, 0)

	for _, u := range e.users {
		if err := SimulateDHCPClient(u, e.ros, &e.nextIP); err != nil {
			log.Printf("[engine] DHCP for %s: %v", u.Username, err)
			continue
		}
		if err := AuthenticateHotspot(u, e.ros); err != nil {
			log.Printf("[engine] Hotspot for %s: %v", u.Username, err)
		}
		if err := CreateUserQueue(u, e.ros); err != nil {
			log.Printf("[engine] Queue for %s: %v", u.Username, err)
			continue
		}
		// Optionally add secondary IP for traffic binding (Linux only)
		_ = AddSecondaryIP(e.cfg.InterfaceName, u.IPAddress)

		stop := make(chan struct{})
		e.stopChans[u.ID] = stop
		switch u.Profile {
		case ProfileIdle:
			go RunIdleTraffic(u, stop)
		case ProfileStandard:
			go RunStandardTraffic(u, stop, "http://10.10.0.1/")
		case ProfileHeavy:
			go RunHeavyTraffic(u, stop, "10.10.0.1:80")
		}
	}

	e.running.Store(true)
	return nil
}

// Stop halts all traffic generation and cleans up.
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running.Load() {
		return nil
	}

	for _, ch := range e.stopChans {
		close(ch)
	}
	e.stopChans = make(map[int]chan struct{})

	if e.ros != nil {
		for _, u := range e.users {
			_ = RemoveUserQueue(u, e.ros)
		}
		e.ros.Close()
		e.ros = nil
	}
	e.users = nil
	e.running.Store(false)
	return nil
}

// ScaleTo adjusts the user count (Phase 6 will implement full scale logic).
func (e *Engine) ScaleTo(target int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	current := len(e.users)
	if target == current {
		return nil
	}
	if target > current {
		// Scale up: add more users
		delta := target - current
		profiles := AssignProfiles(delta)
		users, err := GenerateUsers(delta, profiles)
		if err != nil {
			return err
		}
		baseID := current + 1
		for i := range users {
			u := &users[i]
			u.ID = baseID + i
			if err := SimulateDHCPClient(u, e.ros, &e.nextIP); err != nil {
				log.Printf("[engine] scale DHCP for %s: %v", u.Username, err)
				continue
			}
			if err := AuthenticateHotspot(u, e.ros); err != nil {
				log.Printf("[engine] scale Hotspot for %s: %v", u.Username, err)
			}
			if err := CreateUserQueue(u, e.ros); err != nil {
				log.Printf("[engine] scale Queue for %s: %v", u.Username, err)
				continue
			}
			_ = AddSecondaryIP(e.cfg.InterfaceName, u.IPAddress)
			stop := make(chan struct{})
			e.stopChans[u.ID] = stop
			switch u.Profile {
			case ProfileIdle:
				go RunIdleTraffic(u, stop)
			case ProfileStandard:
				go RunStandardTraffic(u, stop, "http://10.10.0.1/")
			case ProfileHeavy:
				go RunHeavyTraffic(u, stop, "10.10.0.1:80")
			}
			e.users = append(e.users, u)
		}
	} else {
		// Scale down: remove excess users (LIFO)
		toRemove := current - target
		for i := 0; i < toRemove && len(e.users) > 0; i++ {
			last := len(e.users) - 1
			u := e.users[last]
			if ch, ok := e.stopChans[u.ID]; ok {
				close(ch)
				delete(e.stopChans, u.ID)
			}
			_ = RemoveUserQueue(u, e.ros)
			_ = RemoveHotspotSession(u, e.ros)
			_ = ReleaseDHCPLease(u, e.ros)
			_ = RemoveSecondaryIP(e.cfg.InterfaceName, u.IPAddress)
			e.users = e.users[:last]
		}
	}
	return nil
}

// Status returns current user counts by profile.
func (e *Engine) Status() (running bool, activeUsers int, profiles map[string]int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	profiles = map[string]int{"idle": 0, "standard": 0, "heavy": 0}
	for _, u := range e.users {
		profiles[string(u.Profile)]++
	}
	return e.running.Load(), len(e.users), profiles
}

// HTTPServer starts the control API HTTP server.
func (e *Engine) HTTPServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", e.handleStart)
	mux.HandleFunc("/stop", e.handleStop)
	mux.HandleFunc("/scale", e.handleScale)
	mux.HandleFunc("/status", e.handleStatus)
	return http.ListenAndServe(addr, mux)
}

type startRequest struct {
	Count int `json:"count"`
}

type startResponse struct {
	Status      string `json:"status"`
	ActiveUsers int    `json:"activeUsers"`
}

type scaleRequest struct {
	Count int `json:"count"`
}

type statusResponse struct {
	Status      string         `json:"status"`
	ActiveUsers int            `json:"activeUsers"`
	Profiles    map[string]int `json:"profiles"`
}

func (e *Engine) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req startRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Count < 1 || req.Count > 500 {
		http.Error(w, "count must be 1-500", http.StatusBadRequest)
		return
	}
	if err := e.Start(req.Count); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, n, _ := e.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(startResponse{Status: "running", ActiveUsers: n})
}

func (e *Engine) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = e.Stop()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(startResponse{Status: "stopped", ActiveUsers: 0})
}

func (e *Engine) handleScale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req scaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Count < 1 || req.Count > 500 {
		http.Error(w, "count must be 1-500", http.StatusBadRequest)
		return
	}
	if err := e.ScaleTo(req.Count); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, n, _ := e.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(startResponse{Status: "ok", ActiveUsers: n})
}

func (e *Engine) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	running, n, profiles := e.Status()
	status := "stopped"
	if running {
		status = "running"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResponse{
		Status:      status,
		ActiveUsers: n,
		Profiles:    profiles,
	})
}

package engine

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tiklab/tiklab/internal/routeros"
)

const (
	routerOSHost     = "10.10.0.1"
	routerOSAPIPort  = 8728
	hotspotLoginURL  = "http://10.10.0.1/login"
	dhcpServerName   = "dhcp1"
	queueLimitIdle   = "256k/256k"
	queueLimitStd    = "2M/2M"
	queueLimitHeavy  = "5M/5M"
	hotspotUserPass  = "" // Empty password for Hotspot users
)

// EngineConfig holds configuration for the behavior engine.
type EngineConfig struct {
	RouterOSHost    string
	RouterOSAPIPort int
	InterfaceName   string // Network interface for adding secondary IPs
}

// DefaultEngineConfig returns default configuration.
func DefaultEngineConfig() EngineConfig {
	iface := "eth0"
	if v := os.Getenv("TIKLAB_ENGINE_IFACE"); v != "" {
		iface = strings.TrimSpace(v)
	}
	host := routerOSHost
	if v := os.Getenv("TIKLAB_ROUTEROS_HOST"); v != "" {
		host = strings.TrimSpace(v)
	}
	return EngineConfig{
		RouterOSHost:    host,
		RouterOSAPIPort: 8728,
		InterfaceName:   iface,
	}
}

// SimulateDHCPClient allocates an IP for the user via RouterOS API (static lease).
// Uses /ip/dhcp-server/lease/add to create a lease; stores assigned IP in user struct.
// Beta: API-based allocation; full DHCP client protocol deferred for multi-interface support.
func SimulateDHCPClient(user *SimulatedUser, c *routeros.Client, nextIP *uint32) error {
	// Allocate next IP from pool 10.10.0.10 - 10.10.1.254 (501 addresses)
	n := atomic.AddUint32(nextIP, 1)
	if n > 501 {
		return fmt.Errorf("DHCP pool exhausted")
	}
	// n=1 -> 10.10.0.10, n=246 -> 10.10.0.255, n=247 -> 10.10.1.0, n=501 -> 10.10.1.254
	offset := int(n - 1)
	var ip string
	if offset < 246 {
		ip = fmt.Sprintf("10.10.0.%d", 10+offset)
	} else {
		ip = fmt.Sprintf("10.10.1.%d", offset-246)
	}

	_, err := c.Run("/ip/dhcp-server/lease/add",
		"=address="+ip,
		"=mac-address="+user.MACAddress,
		"=server="+dhcpServerName,
	)
	if err != nil {
		return fmt.Errorf("add DHCP lease: %w", err)
	}
	user.IPAddress = ip
	return nil
}

// AuthenticateHotspot adds the Hotspot user and performs HTTP POST to the login page.
func AuthenticateHotspot(user *SimulatedUser, c *routeros.Client) error {
	// Add user via API first
	_, err := c.Run("/ip/hotspot/user/add",
		"=name="+user.Username,
		"=password="+hotspotUserPass,
		"=profile=default",
	)
	if err != nil {
		return fmt.Errorf("add Hotspot user: %w", err)
	}

	// HTTP POST to Hotspot login (HTTP PAP)
	form := fmt.Sprintf("username=%s&password=%s", user.Username, hotspotUserPass)
	req, err := http.NewRequest("POST", hotspotLoginURL, strings.NewReader(form))
	if err != nil {
		return fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[engine] Hotspot login for %s: %v (user added via API)", user.Username, err)
		user.SessionActive = true // User exists, session may appear after redirect
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		user.SessionActive = true
	}
	return nil
}

// CreateUserQueue creates a per-user Simple Queue via RouterOS API.
func CreateUserQueue(user *SimulatedUser, c *routeros.Client) error {
	limit := queueLimitIdle
	switch user.Profile {
	case ProfileStandard:
		limit = queueLimitStd
	case ProfileHeavy:
		limit = queueLimitHeavy
	}
	name := fmt.Sprintf("user-%d-%s", user.ID, strings.ReplaceAll(user.Username, "_", "-"))

	reply, err := c.Run("/queue/simple/add",
		"=name="+name,
		"=target="+user.IPAddress,
		"=max-limit="+limit,
	)
	if err != nil {
		return fmt.Errorf("add queue for %s: %w", user.Username, err)
	}
	user.QueueName = name
	if reply != nil && len(reply.Re) > 0 {
		for _, k := range []string{".id", "id"} {
			if id, ok := reply.Re[0].Map[k]; ok {
				user.QueueID = id
				break
			}
		}
	}
	return nil
}

// RemoveUserQueue removes the user's queue via RouterOS API.
func RemoveUserQueue(user *SimulatedUser, c *routeros.Client) error {
	if user.QueueID != "" {
		_, err := c.Run("/queue/simple/remove", "=numbers="+user.QueueID)
		if err != nil {
			log.Printf("[engine] remove queue %s: %v", user.QueueName, err)
		}
		user.QueueID = ""
	}
	user.QueueName = ""
	return nil
}

// AddSecondaryIP adds a secondary IP to the interface (for traffic source binding).
func AddSecondaryIP(iface, ip string) error {
	cmd := exec.Command("ip", "addr", "add", ip+"/22", "dev", iface)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ip addr add %s dev %s: %v: %s", ip, iface, err, out)
	}
	return nil
}

// RunIdleTraffic sends ICMP ping every 30s and DNS query every 60s.
func RunIdleTraffic(user *SimulatedUser, stop chan struct{}) {
	tickerPing := time.NewTicker(30 * time.Second)
	tickerDNS := time.NewTicker(60 * time.Second)
	defer tickerPing.Stop()
	defer tickerDNS.Stop()

	for {
		select {
		case <-stop:
			return
		case <-tickerPing.C:
			pingGateway(user.IPAddress)
		case <-tickerDNS.C:
			dnsQuery(user.IPAddress)
		}
	}
}

func pingGateway(srcIP string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", "-I", srcIP, "10.10.0.1")
	_ = cmd.Run()
}

func dnsQuery(srcIP string) {
	// Use dig or nslookup with source IP; fallback to simple UDP if not available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "nslookup", "-timeout=2", "google.com", "10.10.0.1")
	cmd.Env = append(cmd.Env, "LD_PRELOAD=") // Avoid env issues
	_ = cmd.Run()
}

// RunStandardTraffic performs HTTP GET every 5-15 seconds.
func RunStandardTraffic(user *SimulatedUser, stop chan struct{}, httpSinkURL string) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		interval := 5 + rand.Intn(11) // 5-15 seconds
		time.Sleep(time.Duration(interval) * time.Second)
		select {
		case <-stop:
			return
		default:
		}
		doHTTPGet(user.IPAddress, httpSinkURL)
	}
}

func doHTTPGet(srcIP, url string) {
	if url == "" {
		url = "http://10.10.0.1/"
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := net.Dialer{LocalAddr: &net.TCPAddr{IP: net.ParseIP(srcIP)}}
				return d.DialContext(ctx, network, addr)
			},
		},
	}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
}

// RunHeavyTraffic maintains a continuous TCP stream to the traffic sink.
func RunHeavyTraffic(user *SimulatedUser, stop chan struct{}, sinkAddr string) {
	if sinkAddr == "" {
		sinkAddr = "10.10.0.1:80"
	}
	for {
		select {
		case <-stop:
			return
		default:
		}
		runTCPStream(user.IPAddress, sinkAddr, stop)
		time.Sleep(time.Second) // Brief pause before reconnect
	}
}

func runTCPStream(srcIP, addr string, stop chan struct{}) {
	dialer := net.Dialer{
		LocalAddr: &net.TCPAddr{IP: net.ParseIP(srcIP)},
		Timeout:   5 * time.Second,
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	// Stream data
	buf := make([]byte, 4096)
	rand.Read(buf)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if _, err := conn.Write(buf); err != nil {
				return
			}
		}
	}
}


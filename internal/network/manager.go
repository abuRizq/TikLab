package network

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const nsPrefix = "sandbox-"

// Manager orchestrates bridge and namespace lifecycle.
type Manager struct {
	Bridge   string
	WorkDir  string
	Profile  Profile
	statePath string
	mu       sync.Mutex
}

// NewManager creates a network manager for the sandbox.
func NewManager(workDir string, profile Profile) *Manager {
	return &Manager{
		Bridge:    DefaultBridge,
		WorkDir:   workDir,
		Profile:   profile,
		statePath: filepath.Join(workDir, "network-state.json"),
	}
}

// Setup creates the bridge and assigns the gateway IP. Call before creating namespaces.
func (m *Manager) Setup() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("network setup requires Linux")
	}
	if !BridgeExists(m.Bridge) {
		if err := CreateBridge(m.Bridge); err != nil {
			return err
		}
	}
	_ = SetBridgeIP(m.Bridge, m.Profile.Gateway+"/24")
	return nil
}

// Teardown removes the bridge and all sandbox namespaces.
func (m *Manager) Teardown() error {
	if runtime.GOOS != "linux" {
		return nil
	}
	names, _ := m.ListNamespaces()
	for _, name := range names {
		ns := m.getNamespaceForDelete(name)
		_ = DeleteNamespace(ns)
	}
	return DeleteBridge(m.Bridge)
}

// ScaleTo sets the number of active client namespaces.
func (m *Manager) ScaleTo(count int) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("scaling requires Linux")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	current, _ := m.ListNamespaces()
	diff := count - len(current)

	if diff > 0 {
		for i := 0; i < diff; i++ {
			idx := len(current) + i
			nsName := m.nsName(idx)
			_, err := CreateNamespace(m.Bridge, nsName, idx)
			if err != nil {
				return fmt.Errorf("create namespace %s: %w", nsName, err)
			}
		}
	} else if diff < 0 {
		toRemove := current[len(current)+diff:]
		for _, name := range toRemove {
			ns := m.getNamespaceForDelete(name)
			_ = DeleteNamespace(ns)
		}
	}
	return nil
}

// ListNamespaces returns names of sandbox namespaces.
func (m *Manager) ListNamespaces() ([]string, error) {
	if runtime.GOOS != "linux" {
		return nil, nil
	}
	out, err := exec.Command("ip", "netns", "list").Output()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// ip netns list format: "id" or "id (id)"
		parts := strings.Fields(line)
		if len(parts) >= 1 && strings.HasPrefix(parts[0], nsPrefix) {
			names = append(names, parts[0])
		}
	}
	return names, nil
}

func (m *Manager) nsName(idx int) string {
	return fmt.Sprintf("%s%04d", nsPrefix, idx)
}

func (m *Manager) getNamespaceForDelete(name string) *Namespace {
	vethHost := fmt.Sprintf("veth-%s-host", name)
	return &Namespace{Name: name, VethHost: vethHost, VethNS: "eth0"}
}

// PersistState saves the current scale count for recovery.
func (m *Manager) PersistState(count int) error {
	return os.WriteFile(m.statePath, []byte(strconv.Itoa(count)), 0644)
}

// LoadState returns the persisted scale count.
func (m *Manager) LoadState() (int, error) {
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

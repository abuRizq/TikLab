package network

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"
)

// Namespace represents a network namespace for a synthetic client.
type Namespace struct {
	Name   string
	VethHost string
	VethNS   string
}

var (
	activeNamespaces   = make(map[string]*Namespace)
	activeNamespacesMu sync.Mutex
)

// CreateNamespace creates a network namespace with a veth pair connected to the bridge.
// Assigns static IP 192.168.88.(10+idx)/24 for traffic generation (DHCP can override later).
func CreateNamespace(bridge, nsName string, idx int) (*Namespace, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("namespaces require Linux (current: %s)", runtime.GOOS)
	}
	vethHost := fmt.Sprintf("veth-%s-host", nsName)
	vethNS := fmt.Sprintf("eth0")

	// Create veth pair
	out, err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethNS).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ip link add veth: %w: %s", err, string(out))
	}

	// Add host end to bridge
	out, err = exec.Command("ip", "link", "set", vethHost, "master", bridge).CombinedOutput()
	if err != nil {
		exec.Command("ip", "link", "del", vethHost).Run()
		return nil, fmt.Errorf("ip link set master: %w: %s", err, string(out))
	}

	// Create namespace and move veth-ns into it
	out, err = exec.Command("ip", "netns", "add", nsName).CombinedOutput()
	if err != nil {
		exec.Command("ip", "link", "set", vethHost, "nomaster").Run()
		exec.Command("ip", "link", "del", vethHost).Run()
		return nil, fmt.Errorf("ip netns add: %w: %s", err, string(out))
	}

	out, err = exec.Command("ip", "link", "set", vethNS, "netns", nsName).CombinedOutput()
	if err != nil {
		exec.Command("ip", "netns", "del", nsName).Run()
		exec.Command("ip", "link", "set", vethHost, "nomaster").Run()
		exec.Command("ip", "link", "del", vethHost).Run()
		return nil, fmt.Errorf("ip link set netns: %w: %s", err, string(out))
	}

	// Bring up host end
	exec.Command("ip", "link", "set", vethHost, "up").Run()

	// In namespace: bring up lo and eth0
	exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", "lo", "up").Run()
	exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethNS, "up").Run()

	// Assign static IP for traffic (192.168.88.10+idx)
	ip := fmt.Sprintf("192.168.88.%d/24", 10+idx)
	exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", ip, "dev", vethNS).Run()

	ns := &Namespace{Name: nsName, VethHost: vethHost, VethNS: vethNS}
	activeNamespacesMu.Lock()
	activeNamespaces[nsName] = ns
	activeNamespacesMu.Unlock()
	return ns, nil
}

// DeleteNamespace removes a namespace and its veth pair.
func DeleteNamespace(ns *Namespace) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	exec.Command("ip", "netns", "del", ns.Name).Run()
	exec.Command("ip", "link", "del", ns.VethHost).Run()
	activeNamespacesMu.Lock()
	delete(activeNamespaces, ns.Name)
	activeNamespacesMu.Unlock()
	return nil
}

// Exec runs a command in the namespace.
func (n *Namespace) Exec(name string, args ...string) *exec.Cmd {
	all := append([]string{"netns", "exec", n.Name, name}, args...)
	return exec.Command("ip", all...)
}

// ExecCombined runs a command and returns combined output.
func (n *Namespace) ExecCombined(name string, args ...string) ([]byte, error) {
	return n.Exec(name, args...).CombinedOutput()
}

// ListActive returns all active namespace names.
func ListActive() []string {
	activeNamespacesMu.Lock()
	defer activeNamespacesMu.Unlock()
	names := make([]string, 0, len(activeNamespaces))
	for n := range activeNamespaces {
		names = append(names, n)
	}
	return names
}

// ActiveCount returns the number of active namespaces.
func ActiveCount() int {
	activeNamespacesMu.Lock()
	defer activeNamespacesMu.Unlock()
	return len(activeNamespaces)
}

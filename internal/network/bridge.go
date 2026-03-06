package network

import (
	"fmt"
	"os/exec"
	"runtime"
)

const DefaultBridge = "br-sandbox"

// CreateBridge creates a Linux bridge. Linux only.
func CreateBridge(name string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("bridge creation requires Linux (current: %s)", runtime.GOOS)
	}
	out, err := exec.Command("ip", "link", "add", "name", name, "type", "bridge").CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link add bridge: %w: %s", err, string(out))
	}
	out, err = exec.Command("ip", "link", "set", name, "up").CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link set up: %w: %s", err, string(out))
	}
	return nil
}

// SetBridgeIP assigns an IP address to the bridge (for traffic targets).
// Idempotent: no-op if IP already assigned.
func SetBridgeIP(bridge, cidr string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	exec.Command("ip", "addr", "add", cidr, "dev", bridge).Run()
	return nil
}

// DeleteBridge removes a bridge and its ports.
func DeleteBridge(name string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	exec.Command("ip", "link", "set", name, "down").Run()
	out, err := exec.Command("ip", "link", "del", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link del bridge: %w: %s", err, string(out))
	}
	return nil
}

// BridgeExists returns true if the bridge exists.
func BridgeExists(name string) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	err := exec.Command("ip", "link", "show", name).Run()
	return err == nil
}

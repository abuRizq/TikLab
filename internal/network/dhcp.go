package network

import (
	"fmt"
	"os/exec"
)

// AcquireDHCP runs a DHCP client in the namespace to obtain an IP via DORA.
// Uses udhcpc (busybox) or dhclient. The CHR must have a DHCP server configured.
func (n *Namespace) AcquireDHCP(iface string) error {
	if iface == "" {
		iface = n.VethNS
	}
	// Try udhcpc first (common in minimal images)
	if path, err := exec.LookPath("udhcpc"); err == nil {
		out, err := n.Exec(path, "-i", iface, "-q", "-t", "3", "-n").CombinedOutput()
		if err != nil {
			return fmt.Errorf("udhcpc: %w: %s", err, string(out))
		}
		return nil
	}
	// Fallback to dhclient
	if path, err := exec.LookPath("dhclient"); err == nil {
		out, err := n.Exec(path, "-v", iface).CombinedOutput()
		if err != nil {
			return fmt.Errorf("dhclient: %w: %s", err, string(out))
		}
		return nil
	}
	return fmt.Errorf("no DHCP client (udhcpc or dhclient) found in PATH")
}

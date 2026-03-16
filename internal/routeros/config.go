package routeros

import (
	"fmt"
)

// DHCP and Hotspot configuration constants per data-model.md
const (
	dhcpPoolName   = "dhcp-pool"
	dhcpPoolRanges = "10.10.0.10-10.10.1.254" // 501 addresses, supports 500-user max
	dhcpAddress    = "10.10.0.1/22"
	dhcpNetwork    = "10.10.0.0/22"
	dhcpGateway    = "10.10.0.1"
	dhcpDNS        = "8.8.8.8"
	dhcpLeaseTime  = "1h"
	dhcpInterface  = "ether1" // CHR default interface
)

// DHCPConfigCommands returns the RouterOS API command sequence for DHCP server setup.
func DHCPConfigCommands() []string {
	return []string{
		"/ip/pool/add",
		"/ip/address/add",
		"/ip/dhcp-server/add",
		"/ip/dhcp-server/network/add",
	}
}

// ConfigureDHCP creates the DHCP server configuration via RouterOS API.
// Creates IP pool 10.10.0.10-10.10.1.254 (501 addresses), assigns 10.10.0.1/22
// to the bridge interface, creates DHCP server with 1h lease, gateway 10.10.0.1, DNS 8.8.8.8.
func ConfigureDHCP(c *Client) error {
	// Create IP pool
	if _, err := c.Run("/ip/pool/add", "=name="+dhcpPoolName, "=ranges="+dhcpPoolRanges); err != nil {
		return fmt.Errorf("add IP pool: %w", err)
	}

	// Assign address to interface
	if _, err := c.Run("/ip/address/add", "=address="+dhcpAddress, "=interface="+dhcpInterface); err != nil {
		return fmt.Errorf("add IP address: %w", err)
	}

	// Create DHCP server
	if _, err := c.Run("/ip/dhcp-server/add",
		"=name=dhcp1",
		"=interface="+dhcpInterface,
		"=address-pool="+dhcpPoolName,
		"=lease-time="+dhcpLeaseTime,
	); err != nil {
		return fmt.Errorf("add DHCP server: %w", err)
	}

	// Add DHCP network
	if _, err := c.Run("/ip/dhcp-server/network/add",
		"=address="+dhcpNetwork,
		"=gateway="+dhcpGateway,
		"=dns-server="+dhcpDNS,
	); err != nil {
		return fmt.Errorf("add DHCP network: %w", err)
	}

	return nil
}

// HotspotConfigCommands returns the RouterOS API command sequence for Hotspot setup.
func HotspotConfigCommands() []string {
	return []string{
		"/ip/hotspot/add",
		"/ip/hotspot/profile/set",
		"/ip/hotspot/user/profile/add",
	}
}

// ConfigureHotspot creates the Hotspot server configuration via RouterOS API.
// Creates Hotspot server on bridge interface, uses DHCP pool, HTTP PAP login, default profile.
func ConfigureHotspot(c *Client) error {
	// Create Hotspot server
	if _, err := c.Run("/ip/hotspot/add",
		"=name=hotspot1",
		"=interface="+dhcpInterface,
		"=address-pool="+dhcpPoolName,
		"=profile=default",
	); err != nil {
		return fmt.Errorf("add Hotspot server: %w", err)
	}

	// Set Hotspot profile for HTTP PAP login
	if _, err := c.Run("/ip/hotspot/profile/set",
		"=numbers=default",
		"=login-by=http-pap",
	); err != nil {
		return fmt.Errorf("set Hotspot profile: %w", err)
	}

	// Create default user profile with no bandwidth limit at Hotspot level
	if _, err := c.Run("/ip/hotspot/user/profile/add",
		"=name=default",
		"=shared-users=1",
	); err != nil {
		return fmt.Errorf("add Hotspot user profile: %w", err)
	}

	return nil
}

// QueueConfigCommands returns the RouterOS API command sequence for queue template setup.
func QueueConfigCommands() []string {
	return []string{
		"/queue/simple/add",
	}
}

// Queue profile limits per behavior type (research.md R8)
const (
	queueLimitIdle     = "256k/256k"
	queueLimitStandard = "2M/2M"
	queueLimitHeavy    = "5M/5M"
)

// ConfigureQueueTemplate defines queue parameters per behavior profile.
// Creates templates: idle 256k/256k, standard 2M/2M, heavy 5M/5M.
// Per-user queues are created by the behavior engine in Phase 5.
func ConfigureQueueTemplate(c *Client) error {
	profiles := []struct {
		name string
		limit string
	}{
		{"profile-idle", queueLimitIdle},
		{"profile-standard", queueLimitStandard},
		{"profile-heavy", queueLimitHeavy},
	}

	for _, p := range profiles {
		if _, err := c.Run("/queue/simple/add",
			"=name="+p.name,
			"=target="+dhcpNetwork,
			"=max-limit="+p.limit,
		); err != nil {
			return fmt.Errorf("add queue %s: %w", p.name, err)
		}
	}

	return nil
}

// ApplyInitialConfig applies DHCP, Hotspot, and queue configuration in sequence.
// Intended to be called after RouterOS boot completes during tiklab start.
func ApplyInitialConfig(c *Client, progress func(msg string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	progress("Configuring DHCP server...")
	if err := ConfigureDHCP(c); err != nil {
		return err
	}

	progress("Configuring Hotspot...")
	if err := ConfigureHotspot(c); err != nil {
		return err
	}

	progress("Configuring queue templates...")
	if err := ConfigureQueueTemplate(c); err != nil {
		return err
	}

	return nil
}

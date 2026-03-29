package routeros

import (
	"fmt"
	"time"

	"github.com/tiklab/tiklab/internal/debug"
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
		"/ip/hotspot/user/profile/set",
	}
}

// ConfigureHotspot creates the Hotspot server configuration via RouterOS API.
// Creates Hotspot server on bridge interface, uses DHCP pool, HTTP PAP login, default profile.
func ConfigureHotspot(c *Client) error {
	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step1_before_add", map[string]interface{}{}, "H1")
	// #endregion

	if _, err := c.Run("/ip/hotspot/add",
		"=name=hotspot1",
		"=interface="+dhcpInterface,
		"=address-pool="+dhcpPoolName,
		"=profile=default",
	); err != nil {
		return fmt.Errorf("add Hotspot server: %w", err)
	}

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step2_after_add", map[string]interface{}{}, "H1")
	// #endregion

	if _, err := c.Run("/ip/hotspot/profile/set",
		"=numbers=default",
		"=login-by=http-pap",
	); err != nil {
		return fmt.Errorf("set Hotspot profile: %w", err)
	}

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step3_after_profile", map[string]interface{}{}, "H1")
	// #endregion

	if _, err := c.Run("/ip/hotspot/user/profile/set",
		"=numbers=default",
		"=shared-users=1",
	); err != nil {
		return fmt.Errorf("set Hotspot user profile: %w", err)
	}

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step4_after_userprofile", map[string]interface{}{}, "H1")
	// #endregion

	// IP binding bypasses hotspot auth for all traffic, keeping API/SSH/Winbox accessible
	// after the hotspot is enabled. Simulated users authenticate via API, not captive portal.
	if _, err := c.Run("/ip/hotspot/ip-binding/add",
		"=address=0.0.0.0/0",
		"=type=bypassed",
		"=server=hotspot1",
		"=comment=TikLab management bypass",
	); err != nil {
		// #region agent log
		debug.Log("routeros/config.go:ConfigureHotspot", "ip_binding_failed", map[string]interface{}{
			"err": err.Error(),
		}, "H1")
		// #endregion
		return fmt.Errorf("add IP binding bypass: %w", err)
	}

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step5_ip_binding_added", map[string]interface{}{}, "H1")
	// #endregion

	// Enabling the hotspot on ether1 disrupts the active API TCP connection
	// (RouterOS reconfigures the interface's network stack, killing existing sessions).
	// Use RouterOS scheduler to enable it asynchronously: the add returns immediately,
	// the script runs after 2s on RouterOS's own scheduler thread, and removes itself.
	if _, err := c.Run("/system/scheduler/add",
		"=name=hs-enable",
		"=on-event=/ip hotspot set hotspot1 disabled=no; /system scheduler remove hs-enable",
		"=interval=2s",
	); err != nil {
		return fmt.Errorf("schedule Hotspot enable: %w", err)
	}

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step6_scheduler_added", map[string]interface{}{}, "H1")
	// #endregion

	time.Sleep(5 * time.Second)

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step6b_before_reconnect", map[string]interface{}{}, "H1")
	// #endregion

	if err := c.Reconnect(); err != nil {
		return fmt.Errorf("reconnect after Hotspot enable: %w", err)
	}

	// #region agent log
	debug.Log("routeros/config.go:ConfigureHotspot", "step6c_reconnected", map[string]interface{}{}, "H1")
	// #endregion

	// #region agent log
	hsPrint, errP := c.Run("/ip/hotspot/print")
	props := map[string]string{}
	printErr := ""
	if errP != nil {
		printErr = errP.Error()
	}
	if errP == nil && hsPrint != nil {
		for _, re := range hsPrint.Re {
			if re.Map["name"] == "hotspot1" {
				for k, v := range re.Map {
					props[k] = v
				}
				break
			}
		}
	}
	debug.Log("routeros/config.go:ConfigureHotspot", "step7_final_state", map[string]interface{}{
		"printErr":         printErr,
		"hotspot1Props":    props,
		"hotspot1Disabled": props["disabled"],
	}, "H1")
	// #endregion

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

// WipeConfig removes all configuration added by TikLab: queues, Hotspot, DHCP, IP, pool.
// Order matters: remove dependents before parents.
// Intended for tiklab reset to restore a clean state.
func WipeConfig(c *Client) error {
	// 1. Remove all simple queues (per-user and templates)
	if err := removeAll(c, "/queue/simple/print", "/queue/simple/remove"); err != nil {
		return fmt.Errorf("remove queues: %w", err)
	}

	// 2. Remove Hotspot IP bindings
	if err := removeAll(c, "/ip/hotspot/ip-binding/print", "/ip/hotspot/ip-binding/remove"); err != nil {
		return fmt.Errorf("remove hotspot ip-bindings: %w", err)
	}

	// 3. Remove Hotspot walled garden IP rules
	if err := removeAll(c, "/ip/hotspot/walled-garden/ip/print", "/ip/hotspot/walled-garden/ip/remove"); err != nil {
		return fmt.Errorf("remove hotspot walled-garden ip: %w", err)
	}

	// 4. Remove Hotspot active sessions
	if err := removeAll(c, "/ip/hotspot/active/print", "/ip/hotspot/active/remove"); err != nil {
		return fmt.Errorf("remove hotspot active: %w", err)
	}

	// 5. Remove Hotspot users
	if err := removeAll(c, "/ip/hotspot/user/print", "/ip/hotspot/user/remove"); err != nil {
		return fmt.Errorf("remove hotspot users: %w", err)
	}

	// 6. Remove Hotspot server (before profiles it references)
	if err := removeAll(c, "/ip/hotspot/print", "/ip/hotspot/remove"); err != nil {
		return fmt.Errorf("remove hotspot: %w", err)
	}

	// 5. Remove Hotspot user profiles (default we added; after server removal)
	if err := removeAll(c, "/ip/hotspot/user/profile/print", "/ip/hotspot/user/profile/remove"); err != nil {
		return fmt.Errorf("remove hotspot profiles: %w", err)
	}

	// 6. Remove DHCP leases
	if err := removeAll(c, "/ip/dhcp-server/lease/print", "/ip/dhcp-server/lease/remove"); err != nil {
		return fmt.Errorf("remove DHCP leases: %w", err)
	}

	// 7. Remove DHCP network
	if err := removeAll(c, "/ip/dhcp-server/network/print", "/ip/dhcp-server/network/remove"); err != nil {
		return fmt.Errorf("remove DHCP network: %w", err)
	}

	// 8. Remove DHCP server
	if err := removeAll(c, "/ip/dhcp-server/print", "/ip/dhcp-server/remove"); err != nil {
		return fmt.Errorf("remove DHCP server: %w", err)
	}

	// 9. Remove IP address on ether1 (10.10.0.1/22)
	if err := removeAll(c, "/ip/address/print", "/ip/address/remove"); err != nil {
		return fmt.Errorf("remove IP address: %w", err)
	}

	// 10. Remove IP pool
	if err := removeAll(c, "/ip/pool/print", "/ip/pool/remove"); err != nil {
		return fmt.Errorf("remove IP pool: %w", err)
	}

	// 11. Remove extra firewall rules (user-added)
	if err := removeAll(c, "/ip/firewall/filter/print", "/ip/firewall/filter/remove"); err != nil {
		return fmt.Errorf("remove firewall rules: %w", err)
	}

	return nil
}

// removeAll prints items at path, collects .id values, and removes each in reverse order.
func removeAll(c *Client, printPath, removePath string) error {
	reply, err := c.Run(printPath)
	if err != nil {
		return err
	}
	if reply == nil || len(reply.Re) == 0 {
		return nil
	}
	ids := make([]string, 0, len(reply.Re))
	for _, re := range reply.Re {
		for _, k := range []string{".id", "id"} {
			if id, ok := re.Map[k]; ok {
				ids = append(ids, id)
				break
			}
		}
	}
	for i := len(ids) - 1; i >= 0; i-- {
		if _, err := c.Run(removePath, "=numbers="+ids[i]); err != nil {
			return err
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

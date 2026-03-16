package routeros

// Config command generation for unit testing.
// Full implementation (ApplyInitialConfig, ConfigureDHCP, etc.) in Phase 4.

// DHCPConfigCommands returns the RouterOS API command sequence for DHCP server setup.
func DHCPConfigCommands() []string {
	return []string{
		"/ip/pool/add",
		"/ip/address/add",
		"/ip/dhcp-server/add",
		"/ip/dhcp-server/network/add",
	}
}

// HotspotConfigCommands returns the RouterOS API command sequence for Hotspot setup.
func HotspotConfigCommands() []string {
	return []string{
		"/ip/hotspot/add",
		"/ip/hotspot/profile/set",
		"/ip/hotspot/user/profile/add",
	}
}

// QueueConfigCommands returns the RouterOS API command sequence for queue template setup.
func QueueConfigCommands() []string {
	return []string{
		"/queue/simple/add",
	}
}

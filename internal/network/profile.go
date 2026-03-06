package network

// Profile defines the isp_small synthetic profile.
type Profile struct {
	Name       string
	UserCount  int
	Subnet     string
	Gateway    string
	DHCPPool   string
	HotspotURL string
}

// ISPSmall returns the hardcoded isp_small profile.
func ISPSmall() Profile {
	return Profile{
		Name:       "isp_small",
		UserCount:  300,
		Subnet:     "192.168.88.0/24",
		Gateway:    "192.168.88.1",
		DHCPPool:   "192.168.88.10-192.168.88.254",
		HotspotURL: "http://192.168.88.1/login",
	}
}

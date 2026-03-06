package network

import (
	"fmt"
	"os/exec"
)

// HotspotAuth simulates Hotspot authentication from within a namespace.
// MAC-cookie: first request triggers redirect to login; cookie identifies session.
// HTTP-PAP: POST credentials to the login URL.
type HotspotAuth struct {
	NS       *Namespace
	LoginURL string
}

// NewHotspotAuth creates a Hotspot auth simulator for a namespace.
func NewHotspotAuth(ns *Namespace, loginURL string) *HotspotAuth {
	return &HotspotAuth{NS: ns, LoginURL: loginURL}
}

// MACCookie simulates MAC-cookie: curl to captive portal to trigger redirect.
func (h *HotspotAuth) MACCookie() error {
	out, err := h.NS.Exec("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-L", "--max-time", "5", "http://captive.router.local/").CombinedOutput()
	if err != nil {
		return fmt.Errorf("curl: %w: %s", err, string(out))
	}
	return nil
}

// HTTPPAP simulates HTTP-PAP: curl with -u to POST credentials.
func (h *HotspotAuth) HTTPPAP(username, password string) error {
	cmd := h.NS.Exec("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-u", username+":"+password, "-X", "POST", h.LoginURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("curl login: %w: %s", err, string(out))
	}
	return nil
}

// CurlAvailable returns true if curl is available in the namespace.
func CurlAvailable(ns *Namespace) bool {
	err := ns.Exec("curl", "--version").Run()
	return err == nil
}

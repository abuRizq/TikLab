package network

import (
	"testing"
)

func TestISPSmall(t *testing.T) {
	p := ISPSmall()
	if p.Name != "isp_small" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.UserCount != 300 {
		t.Errorf("UserCount = %d", p.UserCount)
	}
	if p.Subnet != "192.168.88.0/24" {
		t.Errorf("Subnet = %q", p.Subnet)
	}
	if p.Gateway != "192.168.88.1" {
		t.Errorf("Gateway = %q", p.Gateway)
	}
}

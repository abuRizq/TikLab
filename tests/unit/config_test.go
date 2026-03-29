package unit

import (
	"testing"

	"github.com/tiklab/tiklab/internal/routeros"
)

func TestDHCPConfigCommands(t *testing.T) {
	cmds := routeros.DHCPConfigCommands()
	expected := []string{
		"/ip/pool/add",
		"/ip/address/add",
		"/ip/dhcp-server/add",
		"/ip/dhcp-server/network/add",
	}
	if len(cmds) != len(expected) {
		t.Fatalf("got %d commands, want %d", len(cmds), len(expected))
	}
	for i, c := range cmds {
		if c != expected[i] {
			t.Errorf("cmd[%d]: got %q, want %q", i, c, expected[i])
		}
	}
}

func TestHotspotConfigCommands(t *testing.T) {
	cmds := routeros.HotspotConfigCommands()
	expected := []string{
		"/ip/hotspot/add",
		"/ip/hotspot/profile/set",
		"/ip/hotspot/user/profile/set",
	}
	if len(cmds) != len(expected) {
		t.Fatalf("got %d commands, want %d", len(cmds), len(expected))
	}
	for i, c := range cmds {
		if c != expected[i] {
			t.Errorf("cmd[%d]: got %q, want %q", i, c, expected[i])
		}
	}
}

func TestQueueConfigCommands(t *testing.T) {
	cmds := routeros.QueueConfigCommands()
	expected := []string{"/queue/simple/add"}
	if len(cmds) != len(expected) {
		t.Fatalf("got %d commands, want %d", len(cmds), len(expected))
	}
	for i, c := range cmds {
		if c != expected[i] {
			t.Errorf("cmd[%d]: got %q, want %q", i, c, expected[i])
		}
	}
}

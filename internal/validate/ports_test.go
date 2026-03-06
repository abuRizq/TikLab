package validate

import (
	"testing"
	"time"
)

func TestCheckPort_Closed(t *testing.T) {
	// Use a port that's unlikely to be listening (e.g. 1)
	r := CheckPort("127.0.0.1", 1, 100*time.Millisecond)
	if r.Open {
		t.Error("expected port 1 to be closed")
	}
	if r.Err == nil {
		t.Error("expected error for closed port")
	}
}

func TestCheckAllPorts(t *testing.T) {
	results := CheckAllPorts("127.0.0.1", 100*time.Millisecond)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Name != "API" || results[0].Port != APIPort {
		t.Errorf("first result: name=%s port=%d", results[0].Name, results[0].Port)
	}
	if results[1].Name != "Winbox" || results[1].Port != WinboxPort {
		t.Errorf("second result: name=%s port=%d", results[1].Name, results[1].Port)
	}
	if results[2].Name != "SSH" || results[2].Port != SSHPort {
		t.Errorf("third result: name=%s port=%d", results[2].Name, results[2].Port)
	}
}

func TestCheckAllPorts_EmptyHost(t *testing.T) {
	results := CheckAllPorts("", 100*time.Millisecond)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

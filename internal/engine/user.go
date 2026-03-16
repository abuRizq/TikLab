package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// SimulatedUser represents a virtual network subscriber.
type SimulatedUser struct {
	ID            int
	Username      string
	MACAddress    string
	IPAddress     string
	Profile       Profile
	SessionActive bool
	QueueName     string // RouterOS queue name for teardown
	QueueID       string // RouterOS internal .id for remove
}

// GenerateUsers creates count simulated users with the given profile distribution.
// Ensures uniqueness of username and MAC within the batch.
func GenerateUsers(count int, profiles []Profile) ([]SimulatedUser, error) {
	if count <= 0 {
		return nil, nil
	}
	if len(profiles) < count {
		return nil, fmt.Errorf("insufficient profiles: need %d, got %d", count, len(profiles))
	}

	users := make([]SimulatedUser, count)
	seenUsernames := make(map[string]bool)
	seenMACs := make(map[string]bool)

	for i := 0; i < count; i++ {
		username := generateUniqueUsername(seenUsernames)
		mac := generateUniqueMAC(seenMACs)
		users[i] = SimulatedUser{
			ID:         i + 1,
			Username:   username,
			MACAddress: mac,
			Profile:   profiles[i],
		}
	}
	return users, nil
}

func generateUniqueUsername(seen map[string]bool) string {
	for {
		b := make([]byte, 3)
		_, _ = rand.Read(b)
		username := "guest_" + hex.EncodeToString(b)
		if !seen[username] {
			seen[username] = true
			return username
		}
	}
}

func generateUniqueMAC(seen map[string]bool) string {
	for {
		b := make([]byte, 6)
		_, _ = rand.Read(b)
		// Set locally administered bit (second nibble = 2, 6, A, or E)
		b[0] = (b[0] & 0xFE) | 0x02
		mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
		if !seen[mac] {
			seen[mac] = true
			return mac
		}
	}
}

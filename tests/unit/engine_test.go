package unit

import (
	"testing"

	"github.com/tiklab/tiklab/internal/engine"
)

func TestProfileAssignmentDistribution(t *testing.T) {
	tests := []struct {
		count int
	}{
		{50}, {100}, {200}, {500},
	}
	for _, tt := range tests {
		profiles := engine.AssignProfiles(tt.count)
		if len(profiles) != tt.count {
			t.Errorf("count=%d: got %d profiles", tt.count, len(profiles))
			continue
		}
		idle, standard, heavy := 0, 0, 0
		for _, p := range profiles {
			switch p {
			case engine.ProfileIdle:
				idle++
			case engine.ProfileStandard:
				standard++
			case engine.ProfileHeavy:
				heavy++
			}
		}
		// 40% idle, 45% standard, 15% heavy ±5%
		idlePct := float64(idle) / float64(tt.count) * 100
		standardPct := float64(standard) / float64(tt.count) * 100
		heavyPct := float64(heavy) / float64(tt.count) * 100

		if idlePct < 35 || idlePct > 45 {
			t.Errorf("count=%d: idle %.1f%% outside 35-45%%", tt.count, idlePct)
		}
		if standardPct < 40 || standardPct > 50 {
			t.Errorf("count=%d: standard %.1f%% outside 40-50%%", tt.count, standardPct)
		}
		if heavyPct < 10 || heavyPct > 20 {
			t.Errorf("count=%d: heavy %.1f%% outside 10-20%%", tt.count, heavyPct)
		}
	}
}

func TestUserIdentityUniqueness(t *testing.T) {
	profiles := engine.AssignProfiles(100)
	users, err := engine.GenerateUsers(100, profiles)
	if err != nil {
		t.Fatalf("GenerateUsers: %v", err)
	}
	if len(users) != 100 {
		t.Fatalf("got %d users, want 100", len(users))
	}

	usernames := make(map[string]bool)
	macs := make(map[string]bool)
	for _, u := range users {
		if usernames[u.Username] {
			t.Errorf("duplicate username: %s", u.Username)
		}
		usernames[u.Username] = true
		if macs[u.MACAddress] {
			t.Errorf("duplicate MAC: %s", u.MACAddress)
		}
		macs[u.MACAddress] = true
	}
}

func TestGenerateUsersCountAccuracy(t *testing.T) {
	for _, count := range []int{1, 10, 50, 100, 500} {
		profiles := engine.AssignProfiles(count)
		users, err := engine.GenerateUsers(count, profiles)
		if err != nil {
			t.Fatalf("count=%d: %v", count, err)
		}
		if len(users) != count {
			t.Errorf("count=%d: got %d users", count, len(users))
		}
	}
}

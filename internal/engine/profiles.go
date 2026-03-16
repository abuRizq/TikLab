package engine

import "time"

// Profile represents a behavior profile for simulated users.
type Profile string

const (
	ProfileIdle     Profile = "idle"
	ProfileStandard Profile = "standard"
	ProfileHeavy    Profile = "heavy"
)

// ProfileConfig defines traffic parameters per behavior profile (research.md R8).
type ProfileConfig struct {
	Name             string        // Profile name
	TrafficType      string        // Description of traffic type
	IntervalMin      time.Duration // Min interval between activities
	IntervalMax      time.Duration // Max interval between activities
	ThroughputTarget string        // Target throughput description
}

// ProfileConfigs returns the three behavior profiles per research.md R8.
func ProfileConfigs() map[Profile]ProfileConfig {
	return map[Profile]ProfileConfig{
		ProfileIdle: {
			Name:             "idle",
			TrafficType:      "ICMP ping, DNS query",
			IntervalMin:      30 * time.Second,  // Ping every 30s
			IntervalMax:      60 * time.Second,  // DNS every 60s
			ThroughputTarget: "~1 KB/min",
		},
		ProfileStandard: {
			Name:             "standard",
			TrafficType:      "HTTP GET, small downloads",
			IntervalMin:      5 * time.Second,
			IntervalMax:      15 * time.Second,
			ThroughputTarget: "~50-200 KB/min",
		},
		ProfileHeavy: {
			Name:             "heavy",
			TrafficType:      "Continuous TCP stream",
			IntervalMin:      0,
			IntervalMax:      0,
			ThroughputTarget: "~500 KB-1 MB/min",
		},
	}
}

// AssignProfiles returns a slice of profiles distributed 40% idle, 45% standard, 15% heavy.
// Tolerance: ±5% per spec SC-003.
func AssignProfiles(count int) []Profile {
	if count <= 0 {
		return nil
	}
	profiles := make([]Profile, count)
	idleCount := (count * 40) / 100
	standardCount := (count * 45) / 100
	heavyCount := count - idleCount - standardCount

	i := 0
	for ; i < idleCount; i++ {
		profiles[i] = ProfileIdle
	}
	for ; i < idleCount+standardCount; i++ {
		profiles[i] = ProfileStandard
	}
	for ; i < count; i++ {
		profiles[i] = ProfileHeavy
	}
	return profiles
}

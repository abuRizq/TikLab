package engine

// Profile represents a behavior profile for simulated users.
type Profile string

const (
	ProfileIdle    Profile = "idle"
	ProfileStandard Profile = "standard"
	ProfileHeavy   Profile = "heavy"
)

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

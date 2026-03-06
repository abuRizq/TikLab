package validate

import (
	"time"
)

// ResetResult holds reset timing.
type ResetResult struct {
	Duration time.Duration
	OK       bool
	Err      error
}

// MeasureReset runs a reset function and returns duration. Target: < 2 seconds.
func MeasureReset(resetFn func() error) ResetResult {
	start := time.Now()
	err := resetFn()
	d := time.Since(start)
	return ResetResult{
		Duration: d,
		OK:       err == nil && d < 2*time.Second,
		Err:      err,
	}
}

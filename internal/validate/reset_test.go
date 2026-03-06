package validate

import (
	"errors"
	"testing"
	"time"
)

var errTest = errors.New("test error")

func TestMeasureReset_Fast(t *testing.T) {
	result := MeasureReset(func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if !result.OK {
		t.Errorf("expected OK for fast reset, got duration=%v", result.Duration)
	}
	if result.Err != nil {
		t.Errorf("expected no error: %v", result.Err)
	}
}

func TestMeasureReset_Slow(t *testing.T) {
	result := MeasureReset(func() error {
		time.Sleep(3 * time.Second)
		return nil
	})
	if result.OK {
		t.Error("expected not OK for slow reset (> 2s)")
	}
	if result.Duration < 2*time.Second {
		t.Errorf("duration should be >= 2s, got %v", result.Duration)
	}
}

func TestMeasureReset_Error(t *testing.T) {
	result := MeasureReset(func() error {
		return errTest
	})
	if result.OK {
		t.Error("expected not OK when reset returns error")
	}
	if result.Err != errTest {
		t.Errorf("expected errTest: %v", result.Err)
	}
}


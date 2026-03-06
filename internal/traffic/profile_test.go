package traffic

import (
	"testing"
)

func TestAssignType_40_45_15(t *testing.T) {
	// For 100 namespaces: first 40 idle, next 45 browsing, last 15 heavy
	n := 100
	var idle, browsing, heavy int
	for i := 0; i < n; i++ {
		switch AssignType(i, n) {
		case TypeIdle:
			idle++
		case TypeBrowsing:
			browsing++
		case TypeHeavy:
			heavy++
		}
	}
	if idle != 40 {
		t.Errorf("idle = %d, want 40", idle)
	}
	if browsing != 45 {
		t.Errorf("browsing = %d, want 45", browsing)
	}
	if heavy != 15 {
		t.Errorf("heavy = %d, want 15", heavy)
	}
}

func TestAssignType_SmallN(t *testing.T) {
	// 10 namespaces: 4 idle, 4-5 browsing, 1-2 heavy
	n := 10
	var idle, browsing, heavy int
	for i := 0; i < n; i++ {
		switch AssignType(i, n) {
		case TypeIdle:
			idle++
		case TypeBrowsing:
			browsing++
		case TypeHeavy:
			heavy++
		}
	}
	if idle != 4 {
		t.Errorf("idle = %d, want 4", idle)
	}
	if browsing < 4 || browsing > 5 {
		t.Errorf("browsing = %d, want 4 or 5", browsing)
	}
	if heavy < 1 || heavy > 2 {
		t.Errorf("heavy = %d, want 1 or 2", heavy)
	}
}

func TestAssignType_ZeroN(t *testing.T) {
	if AssignType(0, 0) != TypeIdle {
		t.Error("AssignType(0,0) should return TypeIdle")
	}
}

func TestType_String(t *testing.T) {
	if TypeIdle.String() != "idle" {
		t.Errorf("TypeIdle.String() = %q", TypeIdle.String())
	}
	if TypeBrowsing.String() != "browsing" {
		t.Errorf("TypeBrowsing.String() = %q", TypeBrowsing.String())
	}
	if TypeHeavy.String() != "heavy" {
		t.Errorf("TypeHeavy.String() = %q", TypeHeavy.String())
	}
}

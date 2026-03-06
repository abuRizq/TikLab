package traffic

// Distribution defines the 40/45/15 split for isp_small.
const (
	IdlePct   = 40
	BrowsingPct = 45
	HeavyPct  = 15
)

// AssignType returns the traffic type for namespace at index i of total n.
func AssignType(i, n int) Type {
	if n <= 0 {
		return TypeIdle
	}
	pct := (i * 100) / n
	if pct < IdlePct {
		return TypeIdle
	}
	if pct < IdlePct+BrowsingPct {
		return TypeBrowsing
	}
	return TypeHeavy
}

// Type is the traffic category for a namespace.
type Type int

const (
	TypeIdle Type = iota
	TypeBrowsing
	TypeHeavy
)

func (t Type) String() string {
	switch t {
	case TypeIdle:
		return "idle"
	case TypeBrowsing:
		return "browsing"
	case TypeHeavy:
		return "heavy"
	default:
		return "unknown"
	}
}

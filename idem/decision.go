package idem

type Decision uint8

const (
	DecisionNew Decision = iota + 1
	DecisionReplay
	DecisionConflict
)

func (d Decision) String() string {
	switch d {
	case DecisionNew:
		return "NEW"
	case DecisionReplay:
		return "REPLAY"
	case DecisionConflict:
		return "CONFLICT"
	default:
		return "UNKNOWN"
	}
}

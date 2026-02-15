package idem

type Decision uint8

const (
	DecisionNew Decision = iota + 1
	DecisionReplay
	DecisionConflict
	DecisionInProgress
)

func (d Decision) String() string {
	switch d {
	case DecisionNew:
		return "NEW"
	case DecisionReplay:
		return "REPLAY"
	case DecisionConflict:
		return "CONFLICT"
	case DecisionInProgress:
		return "IN_PROGRESS"
	default:
		return "UNKNOWN"
	}
}

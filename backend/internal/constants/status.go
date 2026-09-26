package constants

// Shared status values are mirrored in frontend/src/types/status.ts. Keeping
// the lists explicit makes state-machine drift visible during code review.

type ContainerState string

const (
	ContainerStateReady      ContainerState = "ready"
	ContainerStateInTransit  ContainerState = "in_transit"
	ContainerStateQuarantine ContainerState = "quarantine"
	ContainerStateCleared    ContainerState = "cleared"
)

var AllContainerState = []string{"ready", "in_transit", "quarantine", "cleared"}

type ExcursionState string

const (
	ExcursionStateOpen     ExcursionState = "open"
	ExcursionStateInReview ExcursionState = "in_review"
	ExcursionStateDecided  ExcursionState = "decided"
	ExcursionStateClosed   ExcursionState = "closed"
)

var AllExcursionState = []string{"open", "in_review", "decided", "closed"}

var TransportContainerTransitions = map[string]map[string]bool{
	"ready":      {"in_transit": true, "quarantine": true},
	"in_transit": {"quarantine": true, "cleared": true, "ready": true},
	"quarantine": {"cleared": true},
	"cleared":    {"quarantine": true},
}

var TemperatureWindowTransitions = map[string]map[string]bool{
	"draft":      {"active": true, "expired": true},
	"active":     {"expired": true, "superseded": true, "draft": true},
	"expired":    {"superseded": true, "active": true},
	"superseded": {"expired": true},
}

var ExcursionEventTransitions = map[string]map[string]bool{
	"open":      {"in_review": true},
	"in_review": {"decided": true, "open": true},
	"decided":   {"closed": true, "in_review": true},
	"closed":    {},
}

var DispositionDecisionTransitions = map[string]map[string]bool{
	"draft":      {"release": true, "quarantine": true, "discard": true},
	"release":    {},
	"quarantine": {},
	"discard":    {},
}

func CanTransition(graph map[string]map[string]bool, from, to string) bool {
	targets, exists := graph[from]
	return exists && targets[to]
}

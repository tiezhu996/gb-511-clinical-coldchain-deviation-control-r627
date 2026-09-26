package constants

import "testing"

func TestTransportContainerTransitionGraph(t *testing.T) {
	if !CanTransition(TransportContainerTransitions, "ready", "in_transit") {
		t.Fatalf("expected ready -> in_transit transition to be allowed")
	}
	if CanTransition(TransportContainerTransitions, "ready", "unknown") {
		t.Fatal("unknown status must never be accepted")
	}
	if CanTransition(TransportContainerTransitions, "quarantine", "in_transit") {
		t.Fatal("quarantined container must require reviewer clearance")
	}
}

func TestDispositionIsFinalAfterIndependentDecision(t *testing.T) {
	for _, state := range []string{"release", "quarantine", "discard"} {
		if CanTransition(DispositionDecisionTransitions, state, "draft") || CanTransition(DispositionDecisionTransitions, state, "release") {
			t.Fatalf("final disposition %s must be immutable", state)
		}
	}
}

func TestExcursionRequiresReviewBeforeDecision(t *testing.T) {
	if CanTransition(ExcursionEventTransitions, "open", "decided") {
		t.Fatal("open excursion must not skip independent review")
	}
	if !CanTransition(ExcursionEventTransitions, "open", "in_review") {
		t.Fatal("open excursion should enter review")
	}
	if !CanTransition(ExcursionEventTransitions, "in_review", "decided") {
		t.Fatal("reviewed excursion should allow a decision")
	}
}

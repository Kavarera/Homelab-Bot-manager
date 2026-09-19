package flow

// ActionType represents what the engine should do after executing a step.
type ActionType string

const (
	// ActionNext transitions to the next specified step.
	ActionNext ActionType = "NEXT"
	// ActionStay stays on the current step (useful for validation retries).
	ActionStay ActionType = "STAY"
	// ActionComplete marks the flow as successfully finished and clears the session.
	ActionComplete ActionType = "COMPLETE"
	// ActionCancel cancels the flow and clears the session.
	ActionCancel ActionType = "CANCEL"
)

// Action represents the outcome returned by a StepHandler.
type Action struct {
	Type     ActionType
	NextStep string
}

// Next creates an action to transition to the next step.
func Next(nextStep string) Action {
	return Action{
		Type:     ActionNext,
		NextStep: nextStep,
	}
}

// Stay creates an action to remain on the current step.
func Stay() Action {
	return Action{
		Type: ActionStay,
	}
}

// Complete creates an action to complete the flow and clear session state.
func Complete() Action {
	return Action{
		Type: ActionComplete,
	}
}

// Cancel creates an action to abort the flow and clear session state.
func Cancel() Action {
	return Action{
		Type: ActionCancel,
	}
}

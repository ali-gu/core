package states

type PracticeState string

const (
	PracticeStateCreated  PracticeState = "CREATED"
	PracticeStateActive   PracticeState = "ACTIVE"
	PracticeStateDisabled PracticeState = "DISABLED"
)

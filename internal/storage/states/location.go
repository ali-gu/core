package states

type LocationState string

const (
	LocationStatePending  LocationState = "PENDING"
	LocationStateActive   LocationState = "ACTIVE"
	LocationStateDisabled LocationState = "DISABLED"
)

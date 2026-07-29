package states

type UserState string

const (
	UserStateActive   UserState = "ACTIVE"
	UserStateInvited  UserState = "INVITED"
	UserStateDisabled UserState = "DISABLED"
)

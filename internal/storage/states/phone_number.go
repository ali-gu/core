package states

type PhoneNumberState string

const (
	PhoneNumberStateReserved PhoneNumberState = "RESERVED"
	PhoneNumberStateActive   PhoneNumberState = "ACTIVE"
	PhoneNumberStateDisabled PhoneNumberState = "DISABLED"
)

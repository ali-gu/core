package phonenumber

type ListAvailablePhoneNumbersParams struct {
	CountryCode string
	AreaCode    string
	Contains    string
}

type CheckAvailableParams struct {
	PhoneNumber string
}

type ReservePhoneNumberParams struct {
	PhoneNumber string
}

type ReservePhoneNumberResult struct {
	ReservationRef string
	PhoneNumber    string
}

type PurchasedPhoneNumber struct {
	ID          string
	PhoneNumber string
	Status      string
}

type AvailablePhoneNumber struct {
	PhoneNumber string
	Reservable  bool
}

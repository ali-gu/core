package phonenumber

import "context"

type IPhoneNumberManager interface {
	ListPurchased(ctx context.Context) ([]PurchasedPhoneNumber, error)
	ListAvailable(ctx context.Context, params ListAvailablePhoneNumbersParams) ([]AvailablePhoneNumber, error)
	CheckAvailable(ctx context.Context, params CheckAvailableParams) error
	Reserve(ctx context.Context, params ReservePhoneNumberParams) (*ReservePhoneNumberResult, error)
}

package telnyx

import (
	"context"
	"fmt"

	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"

	telnyxsdk "github.com/team-telnyx/telnyx-go/v4"
	"github.com/team-telnyx/telnyx-go/v4/packages/param"
)

type Telnyx struct {
	client telnyxsdk.Client
}

var _ phonenumber.IPhoneNumberManager = (*Telnyx)(nil)

func NewTelnyx(client telnyxsdk.Client) *Telnyx {
	return &Telnyx{client: client}
}

func (t *Telnyx) ListPurchased(ctx context.Context) ([]phonenumber.PurchasedPhoneNumber, error) {
	page, err := t.client.PhoneNumbers.List(ctx, telnyxsdk.PhoneNumberListParams{})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("phonenumber: list purchased: %w", err))
	}

	numbers := make([]phonenumber.PurchasedPhoneNumber, len(page.Data))
	for i, n := range page.Data {
		numbers[i] = phonenumber.PurchasedPhoneNumber{
			ID:          n.ID,
			PhoneNumber: n.PhoneNumber,
			Status:      string(n.Status),
		}
	}
	return numbers, nil
}

func (t *Telnyx) ListAvailable(ctx context.Context, params phonenumber.ListAvailablePhoneNumbersParams) ([]phonenumber.AvailablePhoneNumber, error) {
	filter := telnyxsdk.AvailablePhoneNumberListParamsFilter{
		Limit:              param.NewOpt[int64](15),
		ExcludeHeldNumbers: param.NewOpt(true),
	}
	if params.CountryCode != "" {
		filter.CountryCode = param.NewOpt(params.CountryCode)
	}
	if params.AreaCode != "" {
		filter.NationalDestinationCode = param.NewOpt(params.AreaCode)
	}
	if params.Contains != "" {
		filter.PhoneNumber = telnyxsdk.AvailablePhoneNumberListParamsFilterPhoneNumber{
			Contains: param.NewOpt(params.Contains),
		}
	}

	resp, err := t.client.AvailablePhoneNumbers.List(ctx, telnyxsdk.AvailablePhoneNumberListParams{
		Filter: filter,
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("phonenumber: list available: %w", err))
	}

	numbers := make([]phonenumber.AvailablePhoneNumber, len(resp.Data))
	for i, n := range resp.Data {
		numbers[i] = phonenumber.AvailablePhoneNumber{
			PhoneNumber: n.PhoneNumber,
			Reservable:  n.Reservable,
		}
	}
	return numbers, nil
}

func (t *Telnyx) CheckAvailable(ctx context.Context, params phonenumber.CheckAvailableParams) error {
	resp, err := t.client.AvailablePhoneNumbers.List(ctx, telnyxsdk.AvailablePhoneNumberListParams{
		Filter: telnyxsdk.AvailablePhoneNumberListParamsFilter{
			Limit: param.NewOpt[int64](1),
			PhoneNumber: telnyxsdk.AvailablePhoneNumberListParamsFilterPhoneNumber{
				StartsWith: param.NewOpt(params.PhoneNumber),
			},
		},
	})
	if err != nil {
		return rerror.New(fmt.Errorf("phone_number: check available: %w", err))
	}
	if len(resp.Data) == 0 || resp.Data[0].PhoneNumber != params.PhoneNumber {
		return rerror.NewMessage(fmt.Sprintf("phone number %s is not available", params.PhoneNumber), rerror.Validation)
	}
	return nil
}

func (t *Telnyx) Reserve(ctx context.Context, params phonenumber.ReservePhoneNumberParams) (*phonenumber.ReservePhoneNumberResult, error) {
	resp, err := t.client.NumberReservations.New(ctx, telnyxsdk.NumberReservationNewParams{
		PhoneNumbers: []telnyxsdk.ReservedPhoneNumberParam{
			{PhoneNumber: param.NewOpt(params.PhoneNumber)},
		},
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("phone_number: reserve: %w", err))
	}
	if len(resp.Data.PhoneNumbers) == 0 {
		return nil, rerror.NewMessage("phone number reservation returned no numbers", rerror.Internal)
	}

	reserved := resp.Data.PhoneNumbers[0]
	return &phonenumber.ReservePhoneNumberResult{
		ReservationRef: reserved.ID,
		PhoneNumber:    reserved.PhoneNumber,
	}, nil
}

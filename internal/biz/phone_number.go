package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type PhoneNumber struct {
	*Biz

	storage                  storage.Storage
	telnyxPhoneNumberManager phonenumber.IPhoneNumberManager
}

type IPhoneNumber interface {
	Get(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]storage.PhoneNumber, error)
	GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.PhoneNumber, error)
	ListPurchased(ctx context.Context, db storage.DB) ([]phonenumber.PurchasedPhoneNumber, error)
	ListAvailable(ctx context.Context, db storage.DB, input contracts.ListAvailablePhoneNumbersRequest) ([]phonenumber.AvailablePhoneNumber, error)
	Reserve(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.ReservePhoneNumberRequest) (*storage.PhoneNumber, error)
	Delete(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, id ksuid.KSUID) error
	Disable(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, id ksuid.KSUID) (*storage.PhoneNumber, error)
	Activate(ctx context.Context, db storage.DB, input contracts.ActivatePhoneNumberRequest) (*storage.PhoneNumber, error)
}

var _ IPhoneNumber = (*PhoneNumber)(nil)

func (p *PhoneNumber) Get(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]storage.PhoneNumber, error) {
	return p.storage.PhoneNumber.Get(ctx, db, storage.PhoneNumberFilters{PracticeID: &practiceID})
}

func (p *PhoneNumber) GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.PhoneNumber, error) {
	return p.storage.PhoneNumber.GetByID(ctx, db, id)
}

func (p *PhoneNumber) ListPurchased(ctx context.Context, _ storage.DB) ([]phonenumber.PurchasedPhoneNumber, error) {
	return p.telnyxPhoneNumberManager.ListPurchased(ctx)
}

func (p *PhoneNumber) ListAvailable(ctx context.Context, _ storage.DB, input contracts.ListAvailablePhoneNumbersRequest) ([]phonenumber.AvailablePhoneNumber, error) {
	return p.telnyxPhoneNumberManager.ListAvailable(ctx, phonenumber.ListAvailablePhoneNumbersParams{
		CountryCode: input.CountryCode,
		AreaCode:    input.AreaCode,
		Contains:    input.Contains,
	})
}

func (p *PhoneNumber) Reserve(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.ReservePhoneNumberRequest) (*storage.PhoneNumber, error) {
	existing, err := p.storage.PhoneNumber.Get(ctx, db, storage.PhoneNumberFilters{PracticeID: &practiceID, EntityState: ptr.ToPtrOrNil(states.PhoneNumberStateReserved)})
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	if len(existing) > 0 {
		return nil, rerror.NewMessage("practice already has a reserved phone number", rerror.Forbidden)
	}

	phoneNumberID := ksuid.New()
	if err := p.storage.PhoneNumber.Create(ctx, db, storage.PhoneNumber{
		EntityBase:                storage.EntityBase[states.PhoneNumberState]{EntityState: states.PhoneNumberStateReserved},
		ID:                        phoneNumberID,
		PhoneNumberReservationRef: ptr.To(ksuid.New().String()),
		PracticeID:                practiceID,
		PhoneNumber:               input.PhoneNumber,
		CreatedAt:                 time.Now(),
	}); err != nil {
		return nil, rerror.Wrap(err)
	}

	return p.storage.PhoneNumber.GetByID(ctx, db, phoneNumberID)
}

func (p *PhoneNumber) Delete(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, id ksuid.KSUID) error {
	phoneNumber, err := p.storage.PhoneNumber.GetByID(ctx, db, id)
	if err != nil || phoneNumber.PracticeID != practiceID {
		return rerror.NewMessage("phone number not found", rerror.Forbidden)
	}

	if phoneNumber.EntityState == states.PhoneNumberStateActive {
		return rerror.NewMessage("cannot delete an active phone number", rerror.Forbidden)
	}

	if err := p.storage.PhoneNumber.Delete(ctx, db, id); err != nil {
		return rerror.Wrap(err)
	}
	return nil
}

func (p *PhoneNumber) Disable(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, id ksuid.KSUID) (*storage.PhoneNumber, error) {
	phoneNumber, err := p.storage.PhoneNumber.GetByID(ctx, db, id)
	if err != nil || phoneNumber.PracticeID != practiceID {
		return nil, rerror.NewMessage("phone number not found", rerror.Forbidden)
	}
	if phoneNumber.EntityState != states.PhoneNumberStateActive {
		return nil, rerror.NewMessage(fmt.Sprintf("cannot disable phone number: not active (state=%s)", phoneNumber.EntityState), rerror.Forbidden)
	}

	phoneNumber.PhoneNumberToDisabled()
	if err := p.storage.PhoneNumber.Update(ctx, db, *phoneNumber); err != nil {
		return nil, rerror.Wrap(err)
	}

	phoneNumber, err = p.storage.PhoneNumber.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	return phoneNumber, nil
}

func (p *PhoneNumber) Activate(ctx context.Context, db storage.DB, input contracts.ActivatePhoneNumberRequest) (*storage.PhoneNumber, error) {
	phoneNumber, err := p.storage.PhoneNumber.GetByID(ctx, db, input.PhoneNumberID)
	if err != nil {
		return nil, rerror.NewMessage("phone number not found", rerror.Validation)
	}

	phoneNumber.PhoneNumberToActive()
	phoneNumber.PhoneNumberIDRef = &input.PhoneNumberRef
	if err := p.storage.PhoneNumber.Update(ctx, db, *phoneNumber); err != nil {
		return nil, rerror.Wrap(err)
	}

	phoneNumber, err = p.storage.PhoneNumber.GetByID(ctx, db, input.PhoneNumberID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	return phoneNumber, nil
}

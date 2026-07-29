package biz

import (
	"context"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type Location struct {
	*Biz

	storage storage.Storage
}

type ILocation interface {
	Create(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.CreateLocationRequest) (*storage.Location, *storage.EHRS, error)
	Get(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]storage.Location, error)
	GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Location, error)
	Update(ctx context.Context, db storage.DB, id ksuid.KSUID, input contracts.UpdateLocationRequest) (*storage.Location, error)
	Delete(ctx context.Context, db storage.DB, id ksuid.KSUID) error
}

var _ ILocation = (*Location)(nil)

func (l *Location) Create(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.CreateLocationRequest) (*storage.Location, *storage.EHRS, error) {
	practice, err := l.storage.Practice.GetByID(ctx, db, practiceID)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	if practice.Email == nil || practice.Website == nil || practice.ZipCode == nil {
		return nil, nil, rerror.NewMessage("Please update your practice information. Email, Website and ZipCode is required to create a location", rerror.Validation)
	}

	pendingLocations, err := l.storage.Location.Get(ctx, db, storage.LocationFilters{
		PracticeID:  &practiceID,
		EntityState: ptr.To(states.LocationStatePending),
	})
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}
	if len(pendingLocations) > 0 {
		return nil, nil, rerror.NewMessage("a location for this practice is already pending onboarding", rerror.Forbidden)
	}

	subdomain, err := l.existingSubdomainForPractice(ctx, db, practiceID)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	locationID := ksuid.New()
	if err := l.storage.Location.Create(ctx, db, storage.Location{
		EntityBase: storage.EntityBase[states.LocationState]{EntityState: states.LocationStatePending},
		ID:         locationID,
		Address:    input.Address,
		PracticeID: practiceID,
		CreatedAt:  time.Now(),
	}); err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	ehrRecord, err := l.Biz.EHR.Create(ctx, db, contracts.CreateEHRRequest{
		LocationID: locationID,
		Type:       constants.EHRNexHealth,
		EHR:        input.EHR,
		Subdomain:  subdomain,
		Name:       ptr.To(practice.Name),
		Email:      practice.Email,
		ZipCode:    practice.ZipCode,
		Website:    practice.Website,
	})
	if err != nil {
		if delErr := l.storage.Location.Delete(ctx, db, locationID); delErr != nil {
			return nil, nil, rerror.Wrap(delErr)
		}
		return nil, nil, rerror.Wrap(err)
	}

	location, err := l.storage.Location.GetByID(ctx, db, locationID)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	return location, ehrRecord, nil
}

func (l *Location) existingSubdomainForPractice(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) (*string, error) {
	locations, err := l.storage.Location.Get(ctx, db, storage.LocationFilters{
		PracticeID: &practiceID,
	})
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if len(locations) == 0 || locations[0].EHR == nil || locations[0].EHR.Subdomain == "" {
		return nil, nil
	}

	return &locations[0].EHR.Subdomain, nil
}

func (l *Location) Update(ctx context.Context, db storage.DB, id ksuid.KSUID, input contracts.UpdateLocationRequest) (*storage.Location, error) {
	return l.storage.Location.GetByID(ctx, db, id)
}

func (l *Location) Get(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]storage.Location, error) {
	return l.storage.Location.Get(ctx, db, storage.LocationFilters{PracticeID: &practiceID})
}

func (l *Location) GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Location, error) {
	return l.storage.Location.GetByID(ctx, db, id)
}

func (l *Location) Delete(ctx context.Context, db storage.DB, id ksuid.KSUID) error {
	location, err := l.storage.Location.GetByID(ctx, db, id)
	if err != nil {
		return rerror.Wrap(err)
	}
	location.LocationToDisabled()
	return l.storage.Location.Update(ctx, db, *location)
}

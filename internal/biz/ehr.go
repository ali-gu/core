package biz

import (
	"context"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type EHR struct {
	*Biz

	storage storage.Storage
	ehr     ehr.IEHR
}

type IEHR interface {
	Create(ctx context.Context, db storage.DB, input contracts.CreateEHRRequest) (*storage.EHRS, error)
}

var _ IEHR = (*EHR)(nil)

func (e *EHR) Create(ctx context.Context, db storage.DB, input contracts.CreateEHRRequest) (*storage.EHRS, error) {
	onboardingParams := ehr.CreateOnboardingParams{
		InstitutionName:    ptr.From(input.Name),
		InstitutionEmail:   ptr.From(input.Email),
		InstitutionZipCode: ptr.From(input.ZipCode),
		InstitutionWebsite: ptr.From(input.Website),

		Subdomain: ptr.From(input.Subdomain),

		EHRName: input.EHR.String(),
	}

	onboarding, err := e.ehr.CreateOnboarding(ctx, onboardingParams)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	ehrID := ksuid.New()
	if err := e.storage.EHR.Create(ctx, db, storage.EHRS{
		ID:            ehrID,
		Type:          input.Type,
		Subdomain:     onboarding.Subdomain,
		LocationRef:   ptr.ToPtrOrNil(onboarding.LocationID),
		LocationID:    input.LocationID,
		OnboardingURL: onboarding.URL,
		OnboardingID:  onboarding.ID,
		CreatedAt:     time.Now(),
	}); err != nil {
		return nil, rerror.Wrap(err)
	}

	return e.storage.EHR.GetByID(ctx, db, ehrID)
}

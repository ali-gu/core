package biz_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_EHR_Create(t *testing.T) {
	t.Run("success_new_institution_uses_the_given_name", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)

		expiresAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, ehr.CreateOnboardingParams{
			InstitutionName:    "Bright Smiles",
			InstitutionEmail:   "test@nexhealth.com",
			InstitutionZipCode: "10023",
			InstitutionWebsite: "www.example-nexhealth.com",
			EHRName:            "opendental",
		}).Return(&ehr.Onboarding{
			ID:           "onboarding_id",
			Subdomain:    "bright-smiles",
			URL:          "https://app.nexhealth.com/onboardings/onboarding_id",
			URLExpiresAt: expiresAt,
			Status:       "in_progress",
		}, nil)

		ehrRecord, err := bz.EHR.Create(cfg.Ctx, cfg.DB, contracts.CreateEHRRequest{
			LocationID: location.ID,
			Type:       constants.EHRNexHealth,
			Name:       ptr.To("Bright Smiles"),
			Email:      ptr.To("test@nexhealth.com"),
			ZipCode:    ptr.To("10023"),
			Website:    ptr.To("www.example-nexhealth.com"),
			EHR:        constants.NexHealthEHROpenDental,
		})
		require.NoError(t, err)
		require.NotEmpty(t, ehrRecord.ID)
		require.Equal(t, constants.EHRNexHealth, ehrRecord.Type)
		require.Equal(t, "bright-smiles", ehrRecord.Subdomain)
		require.Equal(t, location.ID, ehrRecord.LocationID)
		require.Equal(t, "onboarding_id", ehrRecord.OnboardingID)
		require.Equal(t, "https://app.nexhealth.com/onboardings/onboarding_id", ehrRecord.OnboardingURL)

		updatedLocation, err := cfg.Deps.Storage.Location.GetByID(cfg.Ctx, cfg.DB, location.ID)
		require.NoError(t, err)
		require.Equal(t, states.LocationStatePending, updatedLocation.EntityState)
	})

	t.Run("success_existing_subdomain_skips_the_institution_fields", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, ehr.CreateOnboardingParams{
			Subdomain: "acme-dental",
			EHRName:   "opendental",
		}).Return(&ehr.Onboarding{
			ID:        "onboarding_id",
			Subdomain: "acme-dental",
			Status:    "in_progress",
		}, nil)

		ehrRecord, err := bz.EHR.Create(cfg.Ctx, cfg.DB, contracts.CreateEHRRequest{
			LocationID: location.ID,
			Type:       constants.EHRNexHealth,
			Subdomain:  ptr.To("acme-dental"),
			EHR:        constants.NexHealthEHROpenDental,
		})
		require.NoError(t, err)
		require.Equal(t, "acme-dental", ehrRecord.Subdomain)
	})

	t.Run("error_propagated_from_nexhealth", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, mock.Anything).Return(nil, errors.New("nexhealth unavailable"))

		_, err := bz.EHR.Create(cfg.Ctx, cfg.DB, contracts.CreateEHRRequest{
			LocationID: location.ID,
			Type:       constants.EHRNexHealth,
			Name:       ptr.To("Bright Smiles"),
			Email:      ptr.To("test@nexhealth.com"),
			ZipCode:    ptr.To("10023"),
			Website:    ptr.To("www.example-nexhealth.com"),
			EHR:        constants.NexHealthEHROpenDental,
		})
		require.Error(t, err)

		ehrs, err := cfg.Deps.Storage.EHR.Get(cfg.Ctx, cfg.DB, storage.EHRFilters{LocationID: &location.ID})
		require.NoError(t, err)
		require.Empty(t, ehrs)

		unchangedLocation, err := cfg.Deps.Storage.Location.GetByID(cfg.Ctx, cfg.DB, location.ID)
		require.NoError(t, err)
		require.Equal(t, states.LocationStatePending, unchangedLocation.EntityState)
	})
}

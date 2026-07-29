package biz_test

import (
	"errors"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func validCreateLocationRequest() contracts.CreateLocationRequest {
	return contracts.CreateLocationRequest{
		Address: "1 Infinite Loop",
		EHR:     constants.NexHealthEHROpenDental,
	}
}

func Test_Location_Create(t *testing.T) {
	t.Run("first_location_onboards_the_practice_as_a_new_institution", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz,
			fixtures.WithPracticeName("Bright Smiles"),
			fixtures.WithPracticeEmail("ops@example.com"),
			fixtures.WithPracticeZipCode("95014"),
			fixtures.WithPracticeWebsite("www.example.com"),
		)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, ehr.CreateOnboardingParams{
			InstitutionName:    "Bright Smiles",
			InstitutionEmail:   "ops@example.com",
			InstitutionZipCode: "95014",
			InstitutionWebsite: "www.example.com",
			EHRName:            "opendental",
		}).Return(&ehr.Onboarding{
			ID:        "onboarding_id",
			Subdomain: "bright-smiles",
			Status:    "in_progress",
		}, nil)

		location, ehrRecord, err := bz.Location.Create(cfg.Ctx, cfg.DB, practice.ID, validCreateLocationRequest())
		require.NoError(t, err)
		require.NotEmpty(t, location.ID)
		require.Equal(t, practice.ID, location.PracticeID)
		require.Equal(t, "1 Infinite Loop", location.Address)
		require.Equal(t, states.LocationStatePending, location.EntityState)
		require.False(t, location.CreatedAt.IsZero())

		require.Equal(t, "bright-smiles", ehrRecord.Subdomain)
		require.Equal(t, location.ID, ehrRecord.LocationID)
	})

	t.Run("additional_location_reuses_an_existing_ehr_subdomain", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		existing := fixtures.NewLocation(t, cfg, bz, fixtures.WithLocationPracticeID(practice.ID))
		existingEHRs, err := cfg.Deps.Storage.EHR.Get(cfg.Ctx, cfg.DB, storage.EHRFilters{LocationID: &existing.ID})
		require.NoError(t, err)
		require.Len(t, existingEHRs, 1)

		existing.LocationToActive()
		require.NoError(t, cfg.Deps.Storage.Location.Update(cfg.Ctx, cfg.DB, existing))

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, ehr.CreateOnboardingParams{
			InstitutionName:    practice.Name,
			InstitutionEmail:   *practice.Email,
			InstitutionZipCode: *practice.ZipCode,
			InstitutionWebsite: *practice.Website,
			Subdomain:          existingEHRs[0].Subdomain,
			EHRName:            "opendental",
		}).Return(&ehr.Onboarding{
			ID:        "onboarding_id",
			Subdomain: existingEHRs[0].Subdomain,
			Status:    "in_progress",
		}, nil)

		location, ehrRecord, err := bz.Location.Create(cfg.Ctx, cfg.DB, practice.ID, validCreateLocationRequest())
		require.NoError(t, err)
		require.NotEqual(t, existing.ID, location.ID)
		require.Equal(t, existingEHRs[0].Subdomain, ehrRecord.Subdomain)

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Len(t, locations, 2)
	})

	t.Run("error_from_onboarding_does_not_create_a_location", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, mock.Anything).Return(nil, errors.New("nexhealth unavailable"))

		_, _, err := bz.Location.Create(cfg.Ctx, cfg.DB, practice.ID, validCreateLocationRequest())
		require.Error(t, err)

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Empty(t, locations)
	})

	t.Run("error_when_practice_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, _, err := bz.Location.Create(cfg.Ctx, cfg.DB, ksuid.New(), validCreateLocationRequest())
		require.Error(t, err)
	})

	t.Run("error_when_practice_is_missing_contact_info", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz, fixtures.WithPracticeNoContactInfo())

		_, _, err := bz.Location.Create(cfg.Ctx, cfg.DB, practice.ID, validCreateLocationRequest())
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Validation, rerr.Kind())

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Empty(t, locations)
	})

	t.Run("error_when_a_pending_location_already_exists_for_the_practice", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)
		fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(practice.ID))

		_, _, err := bz.Location.Create(cfg.Ctx, cfg.DB, practice.ID, validCreateLocationRequest())
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Len(t, locations, 1)
	})
}

func Test_Location_Get(t *testing.T) {
	t.Run("success_empty", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, ksuid.New())
		require.NoError(t, err)
		require.Empty(t, locations)
	})

	t.Run("success_orders_active_before_pending_before_disabled", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		pendingLocation := fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(practice.ID))

		activeLocation := fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(activeLocation.ID))

		disabledLocation := fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(practice.ID))
		require.NoError(t, bz.Location.Delete(cfg.Ctx, cfg.DB, disabledLocation.ID))

		fixtures.NewLocation(t, cfg, bz)

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Len(t, locations, 3)
		require.Equal(t, states.LocationStateActive, locations[0].EntityState)
		require.Equal(t, activeLocation.ID, locations[0].ID)
		require.Equal(t, states.LocationStatePending, locations[1].EntityState)
		require.Equal(t, pendingLocation.ID, locations[1].ID)
		require.Equal(t, states.LocationStateDisabled, locations[2].EntityState)
		require.Equal(t, disabledLocation.ID, locations[2].ID)
	})
}

func Test_Location_Update(t *testing.T) {
	t.Run("success_returns_the_location_unchanged", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)

		updated, err := bz.Location.Update(cfg.Ctx, cfg.DB, location.ID, contracts.UpdateLocationRequest{})
		require.NoError(t, err)
		require.Equal(t, location.ID, updated.ID)
		require.Equal(t, location.Address, updated.Address)
		require.Equal(t, location.EntityState, updated.EntityState)
	})

	t.Run("error_when_location_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Location.Update(cfg.Ctx, cfg.DB, ksuid.New(), contracts.UpdateLocationRequest{})
		require.Error(t, err)
	})
}

func Test_Location_Delete(t *testing.T) {
	t.Run("success_disables_the_location", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)

		require.NoError(t, bz.Location.Delete(cfg.Ctx, cfg.DB, location.ID))

		locations, err := bz.Location.Get(cfg.Ctx, cfg.DB, location.PracticeID)
		require.NoError(t, err)
		require.Len(t, locations, 1)
		require.Equal(t, states.LocationStateDisabled, locations[0].EntityState)
		require.NotNil(t, locations[0].DisabledAt)
	})

	t.Run("error_when_location_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		err := bz.Location.Delete(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})
}

func Test_Location_GetByID(t *testing.T) {
	t.Run("success_returns_the_location", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		created := fixtures.NewLocation(t, cfg, bz)

		location, err := bz.Location.GetByID(cfg.Ctx, cfg.DB, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, location.ID)
		require.Equal(t, created.PracticeID, location.PracticeID)
		require.Equal(t, created.Address, location.Address)
	})

	t.Run("success_hydrates_the_joined_ehr", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		created := fixtures.NewLocation(t, cfg, bz)

		ehrs, err := cfg.Deps.Storage.EHR.Get(cfg.Ctx, cfg.DB, storage.EHRFilters{LocationID: &created.ID})
		require.NoError(t, err)
		require.Len(t, ehrs, 1)

		location, err := bz.Location.GetByID(cfg.Ctx, cfg.DB, created.ID)
		require.NoError(t, err)
		require.NotNil(t, location.EHR)
		require.Equal(t, ehrs[0].ID, location.EHR.ID)
		require.Equal(t, ehrs[0].Subdomain, location.EHR.Subdomain)
	})

	t.Run("success_leaves_ehr_nil_when_the_location_has_none", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		created := fixtures.NewPendingLocation(t, cfg, bz)

		location, err := bz.Location.GetByID(cfg.Ctx, cfg.DB, created.ID)
		require.NoError(t, err)
		require.Nil(t, location.EHR)
	})

	t.Run("error_when_the_location_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Location.GetByID(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})
}

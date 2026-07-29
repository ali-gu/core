package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func validLocationRequest() contracts.CreateLocationRequest {
	return contracts.CreateLocationRequest{
		Address: "100 Main St",
		EHR:     constants.NexHealthEHROpenDental,
	}
}

func Test_CreateLocation(t *testing.T) {
	t.Run("201_first_location_onboards_the_practice_as_a_new_institution", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		ehrMock := env.cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, ehr.CreateOnboardingParams{
			InstitutionName:    practice.Name,
			InstitutionEmail:   *practice.Email,
			InstitutionZipCode: *practice.ZipCode,
			InstitutionWebsite: *practice.Website,
			EHRName:            "opendental",
		}).Return(&ehr.Onboarding{
			ID:           "onboarding_id",
			Subdomain:    "acme-dental",
			URL:          "https://app.nexhealth.com/onboardings/onboarding_id",
			URLExpiresAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			Status:       "in_progress",
		}, nil)

		w := env.do(http.MethodPost, "/v1/locations", token, validLocationRequest())
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreateLocationResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "100 Main St", resp.Location.Address)
		require.Equal(t, states.LocationStatePending, resp.Location.Status)
		require.Equal(t, "onboarding_id", resp.EHR.OnboardingID)
		require.Equal(t, "https://app.nexhealth.com/onboardings/onboarding_id", resp.EHR.OnboardingURL)

		locations, err := env.cfg.Deps.Storage.Location.Get(env.cfg.Ctx, env.cfg.DB, storage.LocationFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, locations, 1)
		require.Equal(t, states.LocationStatePending, locations[0].EntityState)

		ehrs, err := env.cfg.Deps.Storage.EHR.Get(env.cfg.Ctx, env.cfg.DB, storage.EHRFilters{LocationID: &locations[0].ID})
		require.NoError(t, err)
		require.Len(t, ehrs, 1)
		require.Equal(t, "acme-dental", ehrs[0].Subdomain)
		require.Nil(t, ehrs[0].LocationRef)
	})

	t.Run("201_additional_location_reuses_an_existing_ehr_subdomain", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		existing := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, env.cfg, env.bz,
			fixtures.WithEHRLocationID(existing.ID),
			fixtures.WithEHRSubdomain("acme-dental"),
		)

		ehrMock := env.cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CreateOnboarding", mock.Anything, ehr.CreateOnboardingParams{
			InstitutionName:    practice.Name,
			InstitutionEmail:   *practice.Email,
			InstitutionZipCode: *practice.ZipCode,
			InstitutionWebsite: *practice.Website,
			Subdomain:          "acme-dental",
			EHRName:            "opendental",
		}).Return(&ehr.Onboarding{
			ID:           "onboarding_id",
			Subdomain:    "acme-dental",
			URL:          "https://app.nexhealth.com/onboardings/onboarding_id",
			URLExpiresAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			Status:       "in_progress",
		}, nil)

		w := env.do(http.MethodPost, "/v1/locations", token, validLocationRequest())
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreateLocationResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "onboarding_id", resp.EHR.OnboardingID)

		locations, err := env.cfg.Deps.Storage.Location.Get(env.cfg.Ctx, env.cfg.DB, storage.LocationFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, locations, 2)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/locations", "", validLocationRequest())
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400_when_a_required_field_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		body := validLocationRequest()
		body.Address = ""
		w := env.do(http.MethodPost, "/v1/locations", token, body)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_the_practice_is_missing_contact_info", func(t *testing.T) {
		env := newE2E(t)
		practice := fixtures.NewPractice(t, env.cfg, env.bz, fixtures.WithPracticeNoContactInfo())
		token := env.authFor(practice.ID)

		w := env.do(http.MethodPost, "/v1/locations", token, validLocationRequest())
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("403_when_a_pending_location_already_exists_for_the_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))

		w := env.do(http.MethodPost, "/v1/locations", token, validLocationRequest())
		require.Equal(t, http.StatusForbidden, w.Code)

		locations, err := env.cfg.Deps.Storage.Location.Get(env.cfg.Ctx, env.cfg.DB, storage.LocationFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, locations, 1)
	})
}

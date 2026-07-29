package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

func Test_GetLocations(t *testing.T) {
	t.Run("200_returns_only_the_authenticated_users_practice_locations", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))

		fixtures.NewLocation(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/locations", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetLocationsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.NotNil(t, resp.Data[0].EHR)
		require.NotEmpty(t, resp.Data[0].EHR.OnboardingID)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/locations", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

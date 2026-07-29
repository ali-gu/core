package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

func Test_GetAgentsAnalytics(t *testing.T) {
	t.Run("200_empty_when_the_practice_has_no_agents", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodGet, "/v1/analytics/agents", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetAgentsAnalyticsResponse
		decodeBody(t, w, &resp)
		require.Empty(t, resp.Data)
		require.Zero(t, resp.Summary.TotalAgents)
	})

	t.Run("200_returns_a_numbers_only_row_per_agent_in_the_users_practice_only", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))
		fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/analytics/agents", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetAgentsAnalyticsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.Equal(t, agentRecord.ID, resp.Data[0].AgentID)
		require.Equal(t, 1, resp.Data[0].Rank)
		require.Equal(t, 1, resp.Summary.TotalAgents)
		require.Equal(t, 1, resp.Summary.ByState["CREATED"])
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/analytics/agents", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func Test_GetLocationsAnalytics(t *testing.T) {
	t.Run("200_groups_locations_by_state_for_the_users_practice_only", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewLocation(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/analytics/locations", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetLocationsAnalyticsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.Equal(t, 1, resp.Summary.TotalLocations)
		require.Equal(t, 1, resp.Summary.ByState["PENDING"])
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/analytics/locations", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func Test_GetPhoneNumbersAnalytics(t *testing.T) {
	t.Run("200_groups_phone_numbers_by_state_for_the_users_practice_only", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))
		fixtures.NewPhoneNumber(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/analytics/phone-numbers", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetPhoneNumbersAnalyticsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.Equal(t, 1, resp.Summary.TotalPhoneNumbers)
		require.Equal(t, 1, resp.Summary.ByState["ACTIVE"])
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/analytics/phone-numbers", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

func Test_GetAgents(t *testing.T) {
	t.Run("200_returns_only_the_authenticated_users_practice_agents", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/agents", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetAgentsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/agents", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

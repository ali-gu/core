package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_DeleteAgent(t *testing.T) {
	t.Run("204_when_the_agent_belongs_to_the_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		w := env.do(http.MethodDelete, "/v1/agents/"+agentRecord.ID.String(), token, nil)
		require.Equal(t, http.StatusNoContent, w.Code)

		agents, err := env.bz.Agent.Get(env.cfg.Ctx, env.cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Empty(t, agents)
	})

	t.Run("403_when_the_agent_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherAgent := fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodDelete, "/v1/agents/"+otherAgent.ID.String(), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodDelete, fmt.Sprintf("/v1/agents/%s", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

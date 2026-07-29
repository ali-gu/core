package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_DisableAgent(t *testing.T) {
	t.Run("200_disables_an_active_agent", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		record, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		record.EntityState = states.AgentStateActive
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, *record))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.DisableAgentResponse
		decodeBody(t, w, &resp)
		require.Equal(t, agentRecord.ID, resp.ID)
		require.Equal(t, states.AgentStateDisabled, resp.Status)
	})

	t.Run("error_disabling_a_non_active_agent", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		record, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		record.EntityState = states.AgentStateCreated
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, *record))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error_disabling_an_already_disabled_agent", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		record, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		record.EntityState = states.AgentStateActive
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, *record))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		w = env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherAgent := fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodPatch, "/v1/agents/"+otherAgent.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPatch, "/v1/agents/"+ksuid.New().String()+"/disable", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

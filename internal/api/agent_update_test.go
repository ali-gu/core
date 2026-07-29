package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_UpdateAgent(t *testing.T) {
	t.Run("200_when_the_agent_belongs_to_the_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))
		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String(), token,
			contracts.UpdateAgentRequest{LocationID: &location.ID})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.UpdateAgentResponse
		decodeBody(t, w, &resp)
		require.NotNil(t, resp.Location)
		require.Equal(t, location.ID, *resp.Location.ID)
	})

	t.Run("200_updates_the_name", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String(), token,
			contracts.UpdateAgentRequest{Name: ptr.To("new_name")})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.UpdateAgentResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "new_name", resp.Name)
	})

	t.Run("403_when_the_location_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String(), token,
			contracts.UpdateAgentRequest{LocationID: &location.ID})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_phone_number_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		reserved, err := env.bz.PhoneNumber.Reserve(env.cfg.Ctx, env.cfg.DB, practice.ID,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550199"})
		require.NoError(t, err)

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String(), token,
			contracts.UpdateAgentRequest{PhoneNumberID: &reserved.ID})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_is_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		record, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		record.EntityState = states.AgentStateActive
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, *record))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String(), token,
			contracts.UpdateAgentRequest{Name: ptr.To("new_name")})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_is_disabled", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		record, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		record.EntityState = states.AgentStateDisabled
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, *record))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String(), token,
			contracts.UpdateAgentRequest{Name: ptr.To("new_name")})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherAgent := fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodPatch, "/v1/agents/"+otherAgent.ID.String(), token, contracts.UpdateAgentRequest{})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_does_not_exist", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPatch, fmt.Sprintf("/v1/agents/%s", ksuid.New()), token, contracts.UpdateAgentRequest{})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPatch, fmt.Sprintf("/v1/agents/%s", ksuid.New()), "", contracts.UpdateAgentRequest{})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

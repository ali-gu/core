package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_ActivateAgent(t *testing.T) {
	t.Run("200_when_the_agent_has_an_active_location_and_phone_number", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentPracticeID(practice.ID), fixtures.WithAgentLocationID(location.ID))

		fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/activate", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.ActivateAgentResponse
		decodeBody(t, w, &resp)
		require.Equal(t, agentRecord.ID, resp.ID)
		require.Equal(t, "ACTIVE", string(resp.Status))
	})

	t.Run("error_when_the_agent_has_no_location_assigned", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/activate", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error_when_the_agent_has_no_phone_number_assigned", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentPracticeID(practice.ID), fixtures.WithAgentLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/activate", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error_when_the_assigned_location_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		agentRecord.AssignLocation(location.ID)
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, agentRecord))

		fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/activate", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error_when_the_assigned_phone_number_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentPracticeID(practice.ID), fixtures.WithAgentLocationID(location.ID))

		reserved, err := env.bz.PhoneNumber.Reserve(env.cfg.Ctx, env.cfg.DB, practice.ID,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550198"})
		require.NoError(t, err)

		agentRecord.AssignPhoneNumber(reserved.ID)
		require.NoError(t, env.cfg.Deps.Storage.Agent.Update(env.cfg.Ctx, env.cfg.DB, agentRecord))

		w := env.do(http.MethodPatch, "/v1/agents/"+agentRecord.ID.String()+"/activate", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherAgent := fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodPatch, "/v1/agents/"+otherAgent.ID.String()+"/activate", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_agent_does_not_exist", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPatch, fmt.Sprintf("/v1/agents/%s/activate", ksuid.New()), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPatch, fmt.Sprintf("/v1/agents/%s/activate", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

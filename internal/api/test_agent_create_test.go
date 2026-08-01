package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

func Test_CreateTestAgent(t *testing.T) {
	testAgent := constants.TestAgentFor(constants.LandscapeTest)

	t.Run("201_creates_a_test_agent_with_its_location_ehr_and_phone_number", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/agents/test-agent", token, nil)
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreateTestAgentResponse
		decodeBody(t, w, &resp)
		require.Equal(t, constants.TestAgentName, resp.Name)
		require.Equal(t, states.AgentStateActive, resp.Status)
		require.Equal(t, testAgent.AgentRef, ptr.From(resp.AgentRef))
		require.Equal(t, testAgent.PhoneNumber, ptr.From(resp.PhoneNumber.PhoneNumber))
		require.Equal(t, constants.TestAgentLocationAddress, ptr.From(resp.Location.Address))

		agentRecord, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, resp.ID)
		require.NoError(t, err)
		require.Equal(t, practice.ID, agentRecord.PracticeID)

		ehrRecord, err := env.cfg.Deps.Storage.EHR.GetByLocationID(env.cfg.Ctx, env.cfg.DB, *agentRecord.LocationID)
		require.NoError(t, err)
		require.Equal(t, testAgent.EHRSubdomain, ehrRecord.Subdomain)
		require.Equal(t, testAgent.EHRLocationRef, ptr.From(ehrRecord.LocationRef))

		phoneNumber, err := env.cfg.Deps.Storage.PhoneNumber.GetByID(env.cfg.Ctx, env.cfg.DB, *agentRecord.PhoneNumberID)
		require.NoError(t, err)
		require.Equal(t, testAgent.PhoneNumberRef, ptr.From(phoneNumber.PhoneNumberIDRef))
	})

	t.Run("201_returns_an_agent_the_practice_can_list", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		require.Equal(t, http.StatusCreated, env.do(http.MethodPost, "/v1/agents/test-agent", token, nil).Code)

		agents, err := env.cfg.Deps.Storage.Agent.Get(env.cfg.Ctx, env.cfg.DB, storage.AgentFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, agents, 1)
		require.Equal(t, constants.TestAgentName, agents[0].Name)
	})

	t.Run("201_for_a_super_admin", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/agents/test-agent", token, nil)
		require.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("403_when_the_practice_already_has_a_test_agent", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		require.Equal(t, http.StatusCreated, env.do(http.MethodPost, "/v1/agents/test-agent", token, nil).Code)

		w := env.do(http.MethodPost, "/v1/agents/test-agent", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("201_for_another_practice_sharing_the_same_telnyx_resources", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		first := env.do(http.MethodPost, "/v1/agents/test-agent", token, nil)
		require.Equal(t, http.StatusCreated, first.Code)

		other := fixtures.NewPractice(t, env.cfg, env.bz)
		second := env.do(http.MethodPost, "/v1/agents/test-agent", env.authFor(other.ID), nil)
		require.Equal(t, http.StatusCreated, second.Code)

		var firstResp, secondResp contracts.CreateTestAgentResponse
		decodeBody(t, first, &firstResp)
		decodeBody(t, second, &secondResp)

		require.NotEqual(t, firstResp.ID, secondResp.ID)
		require.Equal(t, ptr.From(firstResp.AgentRef), ptr.From(secondResp.AgentRef))
		require.Equal(t, ptr.From(firstResp.PhoneNumber.PhoneNumber), ptr.From(secondResp.PhoneNumber.PhoneNumber))
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/agents/test-agent", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

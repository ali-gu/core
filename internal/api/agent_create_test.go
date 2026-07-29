package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

func Test_CreateAgent(t *testing.T) {
	t.Run("201_creates_an_agent_in_the_authenticated_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{Name: "front_desk"})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreateAgentResponse
		decodeBody(t, w, &resp)
		require.NotEmpty(t, resp.ID)
		require.Equal(t, "front_desk", resp.Name)

		agents, err := env.cfg.Deps.Storage.Agent.Get(env.cfg.Ctx, env.cfg.DB, storage.AgentFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, agents, 1)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/agents", "", contracts.CreateAgentRequest{Name: "front_desk"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400_when_the_name_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("201_creates_an_agent_with_an_active_location_and_phone_number", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		phoneNumber := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{
			Name:          "front_desk",
			LocationID:    &location.ID,
			PhoneNumberID: &phoneNumber.ID,
		})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreateAgentResponse
		decodeBody(t, w, &resp)

		agentRecord, err := env.cfg.Deps.Storage.Agent.GetByID(env.cfg.Ctx, env.cfg.DB, resp.ID)
		require.NoError(t, err)
		require.NotNil(t, agentRecord.LocationID)
		require.Equal(t, location.ID, *agentRecord.LocationID)
		require.NotNil(t, agentRecord.PhoneNumberID)
		require.Equal(t, phoneNumber.ID, *agentRecord.PhoneNumberID)
	})

	t.Run("403_when_the_location_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{
			Name:       "front_desk",
			LocationID: &location.ID,
		})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_phone_number_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		reserved, err := env.bz.PhoneNumber.Reserve(env.cfg.Ctx, env.cfg.DB, practice.ID,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550199"})
		require.NoError(t, err)

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{
			Name:          "front_desk",
			PhoneNumberID: &reserved.ID,
		})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_assigning_another_practices_location", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		victim := fixtures.NewPractice(t, env.cfg, env.bz)
		victimLocation := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(victim.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(victimLocation.ID))

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{
			Name:       "front_desk",
			LocationID: &victimLocation.ID,
		})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_assigning_another_practices_phone_number", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		victim := fixtures.NewPractice(t, env.cfg, env.bz)
		victimLocation := fixtures.NewPendingLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(victim.ID))
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(victimLocation.ID))
		victimPhone := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(victimLocation.ID))

		w := env.do(http.MethodPost, "/v1/agents", token, contracts.CreateAgentRequest{
			Name:          "front_desk",
			PhoneNumberID: &victimPhone.ID,
		})
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

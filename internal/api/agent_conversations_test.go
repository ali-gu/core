package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_GetAgentConversations(t *testing.T) {
	t.Run("200_when_the_agent_belongs_to_the_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		w := env.do(http.MethodGet, "/v1/agents/"+agentRecord.ID.String()+"/conversations", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetAgentConversationsResponse
		decodeBody(t, w, &resp)
		require.Empty(t, resp.Data)
	})

	t.Run("200_returns_the_agents_logged_conversations_newest_first", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))
		location := fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		phoneNumber := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		now := time.Now()
		for _, c := range []struct {
			ref       string
			createdAt time.Time
		}{
			{"conv_old", now.Add(-time.Hour)},
			{"conv_new", now},
		} {
			require.NoError(t, env.cfg.Deps.Storage.Conversation.Create(env.cfg.Ctx, env.cfg.DB, storage.Conversation{
				ID:              ksuid.New(),
				AgentID:         agentRecord.ID,
				PhoneNumberID:   phoneNumber.ID,
				LocationID:      location.ID,
				PracticeID:      practice.ID,
				ConversationRef: c.ref,
				Duration:        60000,
				Outcome:         "Appointment booked",
				CreatedAt:       c.createdAt,
			}))
		}

		w := env.do(http.MethodGet, "/v1/agents/"+agentRecord.ID.String()+"/conversations", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetAgentConversationsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 2)

		require.Equal(t, "conv_new", resp.Data[0].ConversationRef)
		require.Equal(t, "conv_old", resp.Data[1].ConversationRef)
		require.Equal(t, int64(60000), resp.Data[0].Duration)
		require.Equal(t, "Appointment booked", resp.Data[0].Outcome)
		require.False(t, resp.Data[0].CreatedAt.IsZero())
	})

	t.Run("403_when_the_agent_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherAgent := fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/agents/"+otherAgent.ID.String()+"/conversations", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, fmt.Sprintf("/v1/agents/%s/conversations", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

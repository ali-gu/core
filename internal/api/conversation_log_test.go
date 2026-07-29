package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

func Test_LogConversation(t *testing.T) {
	t.Run("204_logs_and_derives_outcome", func(t *testing.T) {
		env := newE2E(t)

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz)
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("agent_ref_123"),
		)
		fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", contracts.LogConversationRequest{
			AssistantID:     "agent_ref_123",
			ConversationRef: "conv_123",
			CallDuration:    90000,
			ToolsCalled:     []string{constants.ToolKindBookAppointment.String()},
		})
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Empty(t, w.Body.Bytes())

		conversations, err := env.bz.Conversation.GetByAgent(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Len(t, conversations, 1)
		require.Equal(t, "conv_123", conversations[0].ConversationRef)
		require.Equal(t, "Appointment booked", conversations[0].Outcome)
	})

	t.Run("400_when_conversation_ref_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", contracts.LogConversationRequest{
			AssistantID: "agent_ref_123",
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_body_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "request body is required")
	})
}

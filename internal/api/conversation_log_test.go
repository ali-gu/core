package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/require"
)

const telnyxAssistantID = "assistant-cf1ab171-db2b-48b8-a7b6-57930f547615"

const telnyxInsightEvent = `{
  "event_type": "conversation_insight_result",
  "payload": {
    "conversation_id": "709e61bc-d8ef-48a4-961c-1e3822fa3342",
    "insight_group_id": "f9e76b4c-b1d0-48cf-a20f-c4025238ebda",
    "insight_id": null,
    "insights_instructions": null,
    "metadata": {
      "assistant_id": "assistant-cf1ab171-db2b-48b8-a7b6-57930f547615",
      "assistant_version_id": "20260729T054457360842",
      "call_control_id": "v3:l6_Qa8bP7K4C3fQ4qncF9KeAZ8xi86r6Vv_WmHEpF06_DNKHnAwT8g",
      "call_leg_id": "ab23c410-8b10-11f1-81a1-02420a0d1b20",
      "call_session_id": "ab23b5c4-8b10-11f1-8f72-02420a0d1b20",
      "called_tools": ["hangup", "log_conversation"],
      "from": "52e31a6342bc6dcd@sip.telnyx.com",
      "telnyx_agent_target": "telnyxportal@assistant-cf1ab171-db2b-48b8-a7b6-57930f547615.sip.telnyx.com",
      "telnyx_conversation_channel": "phone_call",
      "telnyx_end_user_target": "52e31a6342bc6dcd@sip.telnyx.com",
      "telnyx_end_user_target_verified": false,
      "to": "telnyxportal@assistant-cf1ab171-db2b-48b8-a7b6-57930f547615.sip.telnyx.com"
    },
    "request_id": "ed836521-878a-4381-adaf-ff2efbefe09c",
    "results": [
      {
        "insight_id": "cd39c4ce-226c-485f-83d8-43ca70df0f35",
        "result": "{\"call_duration\": 20}"
      }
    ],
    "status": "completed",
    "user_id": "7511140e-beb5-4e37-bf28-8c56e264c2f7"
  },
  "record_type": "event"
}`

func Test_LogConversation(t *testing.T) {
	t.Run("204_logs_telnyx_insight_event", func(t *testing.T) {
		env := newE2E(t)

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz)
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef(telnyxAssistantID),
		)
		fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", json.RawMessage(telnyxInsightEvent))
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Empty(t, w.Body.Bytes())

		conversations, err := env.bz.Conversation.GetByAgent(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Len(t, conversations, 1)
		require.Equal(t, "709e61bc-d8ef-48a4-961c-1e3822fa3342", conversations[0].ConversationRef)
		require.Equal(t, int64(20), conversations[0].Duration)
		require.Equal(t, "Info", conversations[0].Outcome)
	})

	t.Run("204_derives_outcome_from_called_tools", func(t *testing.T) {
		env := newE2E(t)

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz)
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef(telnyxAssistantID),
		)
		fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", contracts.LogConversationWebhook{
			Payload: contracts.LogConversationPayload{
				ConversationID: "conv_123",
				Metadata: contracts.LogConversationMetadata{
					AssistantID: telnyxAssistantID,
					CalledTools: []string{"hangup", constants.ToolKindBookAppointment.String()},
				},
				Results: []contracts.LogConversationInsight{{Result: `{"call_duration": 90}`}},
			},
		})
		require.Equal(t, http.StatusNoContent, w.Code)

		conversations, err := env.bz.Conversation.GetByAgent(env.cfg.Ctx, env.cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Len(t, conversations, 1)
		require.Equal(t, "conv_123", conversations[0].ConversationRef)
		require.Equal(t, int64(90), conversations[0].Duration)
		require.Equal(t, "Appointment booked", conversations[0].Outcome)
	})

	t.Run("400_when_conversation_id_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", contracts.LogConversationWebhook{
			Payload: contracts.LogConversationPayload{
				Metadata: contracts.LogConversationMetadata{AssistantID: telnyxAssistantID},
			},
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_assistant_id_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/conversations/log", contracts.LogConversationWebhook{
			Payload: contracts.LogConversationPayload{ConversationID: "conv_123"},
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

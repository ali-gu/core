package contracts_test

import (
	"encoding/json"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/stretchr/testify/require"
)

const telnyxInsightWebhook = `{
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

func Test_LogConversationWebhook_ToRequest(t *testing.T) {
	t.Run("maps_relevant_fields_from_telnyx_event", func(t *testing.T) {
		var webhook contracts.LogConversationWebhook
		require.NoError(t, json.Unmarshal([]byte(telnyxInsightWebhook), &webhook))

		request := webhook.ToRequest()
		require.Equal(t, "assistant-cf1ab171-db2b-48b8-a7b6-57930f547615", request.AssistantID)
		require.Equal(t, "709e61bc-d8ef-48a4-961c-1e3822fa3342", request.ConversationRef)
		require.Equal(t, int64(20), request.CallDuration)
		require.Equal(t, []string{"hangup", "log_conversation"}, request.ToolsCalled)
	})

	t.Run("picks_call_duration_from_the_insight_that_carries_it", func(t *testing.T) {
		webhook := contracts.LogConversationWebhook{
			Payload: contracts.LogConversationPayload{
				ConversationID: "conv_123",
				Metadata:       contracts.LogConversationMetadata{AssistantID: "assistant-123"},
				Results: []contracts.LogConversationInsight{
					{Result: "caller sounded happy"},
					{Result: `{"sentiment": "positive"}`},
					{Result: `{"call_duration": 42.7}`},
				},
			},
		}

		require.Equal(t, int64(42), webhook.ToRequest().CallDuration)
	})

	t.Run("call_duration_is_zero_when_no_insight_reports_it", func(t *testing.T) {
		webhook := contracts.LogConversationWebhook{
			Payload: contracts.LogConversationPayload{
				ConversationID: "conv_123",
				Metadata:       contracts.LogConversationMetadata{AssistantID: "assistant-123"},
				Results:        []contracts.LogConversationInsight{{Result: `{"sentiment": "positive"}`}},
			},
		}

		require.Zero(t, webhook.ToRequest().CallDuration)
	})
}

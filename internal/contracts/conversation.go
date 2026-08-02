package contracts

import "encoding/json"

type LogConversationWebhook struct {
	Payload LogConversationPayload `json:"payload"`
}

type LogConversationPayload struct {
	ConversationID string                   `json:"conversation_id" binding:"required"`
	Metadata       LogConversationMetadata  `json:"metadata"`
	Results        []LogConversationInsight `json:"results"`
}

type LogConversationMetadata struct {
	AssistantID string   `json:"assistant_id" binding:"required"`
	CalledTools []string `json:"called_tools"`
}

type LogConversationInsight struct {
	Result string `json:"result"`
}

type LogConversationRequest struct {
	AssistantID     string
	ConversationRef string
	CallDuration    int64
	ToolsCalled     []string
}

func (w LogConversationWebhook) ToRequest() LogConversationRequest {
	return LogConversationRequest{
		AssistantID:     w.Payload.Metadata.AssistantID,
		ConversationRef: w.Payload.ConversationID,
		CallDuration:    w.Payload.CallDuration(),
		ToolsCalled:     w.Payload.Metadata.CalledTools,
	}
}

func (p LogConversationPayload) CallDuration() int64 {
	for _, insight := range p.Results {
		var parsed struct {
			CallDuration *float64 `json:"call_duration"`
		}
		if err := json.Unmarshal([]byte(insight.Result), &parsed); err != nil || parsed.CallDuration == nil {
			continue
		}
		return int64(*parsed.CallDuration)
	}
	return 0
}

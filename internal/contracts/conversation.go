package contracts

type LogConversationRequest struct {
	AssistantID     string   `json:"assistant_id" binding:"required"`
	ConversationRef string   `json:"conversation_ref" binding:"required"`
	CallDuration    int64    `json:"call_duration"`
	ToolsCalled     []string `json:"tools_called"`
}

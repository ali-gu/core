package contracts

import (
	"time"

	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
)

type AgentURI struct {
	AgentID ksuid.KSUID `uri:"agent_id,parser=encoding.TextUnmarshaler"`
}

type AgentConversationURI struct {
	AgentID        ksuid.KSUID `uri:"agent_id,parser=encoding.TextUnmarshaler"`
	ConversationID string      `uri:"conversation_id"`
}

type CreateAgentRequest struct {
	Name          string       `json:"name" binding:"required"`
	LocationID    *ksuid.KSUID `json:"location_id"`
	PhoneNumberID *ksuid.KSUID `json:"phone_number_id"`
}

type CreateAgentResponse struct {
	ID        ksuid.KSUID `json:"id"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
}

type Agent struct {
	ID          ksuid.KSUID       `json:"id"`
	Status      states.AgentState `json:"status"`
	Name        string            `json:"name"`
	CreatedAt   time.Time         `json:"created_at"`
	Location    *AgentLocation    `json:"location"`
	PhoneNumber *AgentPhoneNumber `json:"phone_number"`
}

type AgentLocation struct {
	ID      *ksuid.KSUID `json:"id"`
	Address *string      `json:"address"`
}

type AgentPhoneNumber struct {
	ID          *ksuid.KSUID `json:"id"`
	PhoneNumber *string      `json:"phone_number"`
}

type GetAgentsResponse struct {
	Data []Agent `json:"data"`
}

type UpdateAgentRequest struct {
	LocationID    *ksuid.KSUID `json:"location_id"`
	PhoneNumberID *ksuid.KSUID `json:"phone_number_id"`
	Name          *string      `json:"name"`
}

type UpdateAgentResponse struct {
	ID          ksuid.KSUID       `json:"id"`
	Status      states.AgentState `json:"status"`
	Name        string            `json:"name"`
	PhoneNumber *AgentPhoneNumber `json:"phone_number"`
	Location    *AgentLocation    `json:"location"`
	CreatedAt   time.Time         `json:"created_at"`
}

type ActivateAgentResponse struct {
	ID          ksuid.KSUID       `json:"id"`
	Status      states.AgentState `json:"status"`
	Name        string            `json:"name"`
	PhoneNumber *AgentPhoneNumber `json:"phone_number"`
	Location    *AgentLocation    `json:"location"`
	CreatedAt   time.Time         `json:"created_at"`
}

type DisableAgentResponse struct {
	ID          ksuid.KSUID       `json:"id"`
	Status      states.AgentState `json:"status"`
	Name        string            `json:"name"`
	PhoneNumber *AgentPhoneNumber `json:"phone_number"`
	Location    *AgentLocation    `json:"location"`
	CreatedAt   time.Time         `json:"created_at"`
}

type AgentAnalyticsMessage struct {
	Role      string                   `json:"role"`
	Text      string                   `json:"text"`
	SentAt    time.Time                `json:"sent_at"`
	ToolCalls []AgentAnalyticsToolCall `json:"tool_calls"`
}

type AgentAnalyticsToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type AgentAnalyticsRecording struct {
	ID             string `json:"id"`
	DurationMillis int64  `json:"duration_millis"`
	MP3URL         string `json:"mp3_url"`
	WavURL         string `json:"wav_url"`
	StartedAt      string `json:"started_at"`
	EndedAt        string `json:"ended_at"`
}

type AgentConversationSummary struct {
	ConversationRef string    `json:"conversation_ref"`
	Duration        int64     `json:"duration"`
	Outcome         string    `json:"outcome"`
	CreatedAt       time.Time `json:"created_at"`
}

type GetAgentConversationsResponse struct {
	Data []AgentConversationSummary `json:"data"`
}

type AgentConversationDetail struct {
	Messages []AgentAnalyticsMessage   `json:"messages"`
	Audio    []AgentAnalyticsRecording `json:"audio"`
	Insights map[string]string         `json:"insights"`
}

type GetAgentConversationResponse struct {
	Data AgentConversationDetail `json:"data"`
}

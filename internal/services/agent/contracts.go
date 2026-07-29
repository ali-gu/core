package agent

import (
	"errors"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
)

var ErrConversationNotFound = errors.New("conversation not found")

type CreateAgentParams struct {
	Name             string
	Model            constants.TelnyxModel
	Instructions     string
	Greeting         string
	PhoneNumberIDRef string
	ToolIDs          []string
}

type CreateAgentResult struct {
	ID           string
	Name         string
	Model        constants.TelnyxModel
	Instructions string
	CreatedAt    time.Time
}

type DeleteAgentParams struct {
	AgentRef string
}

type GetAnalyticsParams struct {
	AgentRef string
}

type GetConversationParams struct {
	AgentRef       string
	ConversationID string
}

type GetRecordingsParams struct {
	CallControlID string
}

type ConversationAnalytics struct {
	ConversationID string
	Status         string
	CreatedAt      time.Time
	LastMessageAt  time.Time
	Metadata       map[string]string
	Insights       map[string]string
	Messages       []ConversationMessage
	Recordings     []ConversationRecording
}

type ConversationMessage struct {
	Role      string
	Text      string
	SentAt    time.Time
	ToolCalls []ConversationToolCall
}

type ConversationToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type ConversationRecording struct {
	ID             string
	DurationMillis int64
	MP3URL         string
	WavURL         string
	StartedAt      string
	EndedAt        string
}

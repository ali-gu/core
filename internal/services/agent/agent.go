package agent

import (
	"context"
)

type IAgent interface {
	Create(ctx context.Context, params CreateAgentParams) (*CreateAgentResult, error)
	Delete(ctx context.Context, params DeleteAgentParams) error
	GetAnalytics(ctx context.Context, params GetAnalyticsParams) ([]ConversationAnalytics, error)
	GetConversation(ctx context.Context, params GetConversationParams) (*ConversationAnalytics, error)
	GetRecordings(ctx context.Context, params GetRecordingsParams) ([]ConversationRecording, error)
}

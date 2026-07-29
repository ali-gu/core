package biz

import (
	"context"
	"sort"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
	"golang.org/x/sync/errgroup"
)

type Analytics struct {
	*Biz

	storage     storage.Storage
	telnyxAgent agent.IAgent
}

type IAnalytics interface {
	GetAgentsAnalytics(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]AgentPerformance, error)
}

var _ IAnalytics = (*Analytics)(nil)

type AgentPerformance struct {
	Agent                storage.Agent
	ConversationCount    int
	BookingsMade         int
	LongestConversation  time.Duration
	ShortestConversation time.Duration
	LastConversationAt   *time.Time
}

func (p AgentPerformance) ConversionRate() float64 {
	if p.ConversationCount == 0 {
		return 0
	}
	return float64(p.BookingsMade) / float64(p.ConversationCount)
}

func (a *Analytics) GetAgentsAnalytics(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]AgentPerformance, error) {
	agents, err := a.storage.Agent.Get(ctx, db, storage.AgentFilters{PracticeID: &practiceID})
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	performances := make([]AgentPerformance, len(agents))
	g, gctx := errgroup.WithContext(ctx)
	for i, agentRecord := range agents {
		g.Go(func() error {
			conversations, err := a.telnyxAgent.GetAnalytics(gctx, agent.GetAnalyticsParams{AgentRef: ptr.From(agentRecord.AgentRef)})
			if err != nil {
				return rerror.Wrap(err)
			}
			performances[i] = summarizeAgentPerformance(agentRecord, conversations)
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, rerror.Wrap(err)
	}

	sort.SliceStable(performances, func(i, j int) bool {
		if performances[i].BookingsMade != performances[j].BookingsMade {
			return performances[i].BookingsMade > performances[j].BookingsMade
		}
		return performances[i].ConversationCount > performances[j].ConversationCount
	})

	return performances, nil
}

func summarizeAgentPerformance(agentRecord storage.Agent, conversations []agent.ConversationAnalytics) AgentPerformance {
	performance := AgentPerformance{
		Agent:             agentRecord,
		ConversationCount: len(conversations),
	}

	for i, conversation := range conversations {
		if conversationBookedAppointment(conversation.Messages) {
			performance.BookingsMade++
		}

		duration := conversation.LastMessageAt.Sub(conversation.CreatedAt)
		if i == 0 || duration > performance.LongestConversation {
			performance.LongestConversation = duration
		}
		if i == 0 || duration < performance.ShortestConversation {
			performance.ShortestConversation = duration
		}

		if performance.LastConversationAt == nil || conversation.CreatedAt.After(*performance.LastConversationAt) {
			createdAt := conversation.CreatedAt
			performance.LastConversationAt = &createdAt
		}
	}

	return performance
}

func conversationBookedAppointment(messages []agent.ConversationMessage) bool {
	for _, message := range messages {
		for _, toolCall := range message.ToolCalls {
			if toolCall.Name == constants.ToolKindBookAppointment.String() {
				return true
			}
		}
	}
	return false
}

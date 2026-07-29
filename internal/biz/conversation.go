package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type Conversation struct {
	*Biz

	storage     storage.Storage
	telnyxAgent agent.IAgent
}

type IConversation interface {
	LogConversation(ctx context.Context, db storage.DB, input contracts.LogConversationRequest) error
	GetByAgent(ctx context.Context, db storage.DB, agentID ksuid.KSUID) ([]storage.Conversation, error)
	GetConversation(ctx context.Context, db storage.DB, id ksuid.KSUID, conversationID string) (*agent.ConversationAnalytics, error)
}

var _ IConversation = (*Conversation)(nil)

func (c *Conversation) LogConversation(ctx context.Context, db storage.DB, input contracts.LogConversationRequest) error {
	agentRecord, err := c.storage.Agent.GetByAgentRef(ctx, db, input.AssistantID)
	if err != nil {
		return rerror.Wrap(err)
	}
	if agentRecord.LocationID == nil {
		return rerror.NewMessage(fmt.Sprintf("agent %s has no location assigned", agentRecord.ID), rerror.Validation)
	}
	if agentRecord.PhoneNumberID == nil {
		return rerror.NewMessage(fmt.Sprintf("agent %s has no phone number assigned", agentRecord.ID), rerror.Validation)
	}

	if err := c.storage.Conversation.Create(ctx, db, storage.Conversation{
		ID:              ksuid.New(),
		AgentID:         agentRecord.ID,
		PhoneNumberID:   *agentRecord.PhoneNumberID,
		LocationID:      *agentRecord.LocationID,
		PracticeID:      agentRecord.PracticeID,
		ConversationRef: input.ConversationRef,
		Duration:        input.CallDuration,
		Outcome:         SummarizeConversationOutcome(input.ToolsCalled),
		CreatedAt:       time.Now(),
	}); err != nil {
		return rerror.Wrap(err)
	}

	return nil
}

func (c *Conversation) GetByAgent(ctx context.Context, db storage.DB, agentID ksuid.KSUID) ([]storage.Conversation, error) {
	return c.storage.Conversation.Get(ctx, db, storage.ConversationFilters{AgentID: &agentID})
}

func (c *Conversation) GetConversation(ctx context.Context, db storage.DB, id ksuid.KSUID, conversationID string) (*agent.ConversationAnalytics, error) {
	agentRecord, err := c.storage.Agent.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	detail, err := c.telnyxAgent.GetConversation(ctx, agent.GetConversationParams{
		AgentRef:       ptr.From(agentRecord.AgentRef),
		ConversationID: conversationID,
	})
	if err != nil {
		if errors.Is(err, agent.ErrConversationNotFound) {
			return nil, rerror.NewMessage("conversation not found", rerror.Forbidden)
		}
		return nil, rerror.Wrap(err)
	}

	recordings, err := c.telnyxAgent.GetRecordings(ctx, agent.GetRecordingsParams{CallControlID: detail.Metadata["call_control_id"]})
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	detail.Recordings = recordings

	return detail, nil
}

func SummarizeConversationOutcome(toolsCalled []string) string {
	var bookedAppointment, foundAppointment, cancelAppointment bool
	for _, name := range toolsCalled {
		switch name {
		case constants.ToolKindBookAppointment.String():
			bookedAppointment = true
		case constants.ToolKindFindAppointment.String():
			foundAppointment = true
		case constants.ToolKindCancelAppointment.String():
			cancelAppointment = true
		}
	}

	switch {
	case bookedAppointment:
		return "Appointment booked"
	case foundAppointment:
		return "Appointment found"
	case cancelAppointment:
		return "Appointment cancelled"
	default:
		return "Info"
	}
}

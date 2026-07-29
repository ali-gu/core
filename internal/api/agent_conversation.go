package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetAgentConversationHandler struct {
	*API

	uri contracts.AgentConversationURI
}

func (h *GetAgentConversationHandler) IsWrite() bool {
	return false
}

func (h *GetAgentConversationHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *GetAgentConversationHandler) Permissions(ctx *gin.Context, db storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	agentRecord, err := h.Biz.Agent.GetByID(ctx, db, h.uri.AgentID)
	if err != nil || agentRecord.PracticeID != user.PracticeID {
		return rerror.NewMessage("forbidden", rerror.Forbidden)
	}
	return nil
}

func (h *GetAgentConversationHandler) Handle(ctx *gin.Context, db storage.DB) error {
	conversation, err := h.Biz.Conversation.GetConversation(ctx, db, h.uri.AgentID, h.uri.ConversationID)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := contracts.AgentConversationDetail{
		Messages: mapConversationMessages(conversation.Messages),
		Audio:    mapConversationRecordings(conversation.Recordings),
		Insights: conversation.Insights,
	}

	ctx.JSON(http.StatusOK, contracts.GetAgentConversationResponse{Data: data})
	return nil
}

func mapConversationMessages(messages []agent.ConversationMessage) []contracts.AgentAnalyticsMessage {
	result := make([]contracts.AgentAnalyticsMessage, len(messages))
	for j, message := range messages {
		toolCalls := make([]contracts.AgentAnalyticsToolCall, len(message.ToolCalls))
		for k, toolCall := range message.ToolCalls {
			toolCalls[k] = contracts.AgentAnalyticsToolCall{
				ID:        toolCall.ID,
				Name:      toolCall.Name,
				Arguments: toolCall.Arguments,
			}
		}

		result[j] = contracts.AgentAnalyticsMessage{
			Role:      message.Role,
			Text:      message.Text,
			SentAt:    message.SentAt,
			ToolCalls: toolCalls,
		}
	}
	return result
}

func mapConversationRecordings(recordings []agent.ConversationRecording) []contracts.AgentAnalyticsRecording {
	result := make([]contracts.AgentAnalyticsRecording, len(recordings))
	for j, r := range recordings {
		result[j] = contracts.AgentAnalyticsRecording{
			ID:             r.ID,
			DurationMillis: r.DurationMillis,
			MP3URL:         r.MP3URL,
			WavURL:         r.WavURL,
			StartedAt:      r.StartedAt,
			EndedAt:        r.EndedAt,
		}
	}
	return result
}

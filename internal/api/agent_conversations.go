package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetAgentConversationsHandler struct {
	*API

	uri contracts.AgentURI
}

func (h *GetAgentConversationsHandler) IsWrite() bool {
	return false
}

func (h *GetAgentConversationsHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *GetAgentConversationsHandler) Permissions(ctx *gin.Context, db storage.DB) error {
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

func (h *GetAgentConversationsHandler) Handle(ctx *gin.Context, db storage.DB) error {
	conversations, err := h.Biz.Conversation.GetByAgent(ctx, db, h.uri.AgentID)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.AgentConversationSummary, len(conversations))
	for i, item := range conversations {
		data[i] = contracts.AgentConversationSummary{
			ConversationRef: item.ConversationRef,
			Duration:        item.Duration,
			Outcome:         item.Outcome,
			CreatedAt:       item.CreatedAt,
		}
	}

	ctx.JSON(http.StatusOK, contracts.GetAgentConversationsResponse{Data: data})
	return nil
}

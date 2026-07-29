package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type DeleteAgentHandler struct {
	*API

	uri contracts.AgentURI
}

func (h *DeleteAgentHandler) IsWrite() bool {
	return true
}

func (h *DeleteAgentHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *DeleteAgentHandler) Permissions(ctx *gin.Context, db storage.DB) error {
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

func (h *DeleteAgentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	if err := h.Biz.Agent.Delete(ctx, db, h.uri.AgentID); err != nil {
		return rerror.Wrap(err)
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

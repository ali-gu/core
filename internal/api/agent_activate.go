package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type ActivateAgentHandler struct {
	*API

	uri contracts.AgentURI
}

func (h *ActivateAgentHandler) IsWrite() bool {
	return true
}

func (h *ActivateAgentHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *ActivateAgentHandler) Permissions(ctx *gin.Context, db storage.DB) error {
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

func (h *ActivateAgentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	updated, err := h.Biz.Agent.Activate(ctx, db, h.uri.AgentID)
	if err != nil {
		return rerror.Wrap(err)
	}

	location := ptr.From(updated.Location)
	phoneNumber := ptr.From(updated.PhoneNumber)

	ctx.JSON(http.StatusOK, contracts.ActivateAgentResponse{
		ID:     updated.ID,
		Status: updated.EntityState,
		Name:   updated.Name,
		Location: &contracts.AgentLocation{
			ID:      ptr.ToPtrOrNil(location.ID),
			Address: ptr.ToPtrOrNil(location.Address),
		},
		PhoneNumber: &contracts.AgentPhoneNumber{
			ID:          ptr.ToPtrOrNil(phoneNumber.ID),
			PhoneNumber: ptr.ToPtrOrNil(phoneNumber.PhoneNumber),
		},
		CreatedAt: updated.CreatedAt,
	})
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type UpdateAgentHandler struct {
	*API

	uri contracts.AgentURI
	req contracts.UpdateAgentRequest
}

func (h *UpdateAgentHandler) IsWrite() bool {
	return true
}

func (h *UpdateAgentHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	if err := ctx.ShouldBindUri(&h.uri); err != nil {
		return bindErr(err)
	}
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *UpdateAgentHandler) Permissions(ctx *gin.Context, db storage.DB) error {
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

func (h *UpdateAgentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	updated, err := h.Biz.Agent.Update(ctx, db, h.uri.AgentID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	location := ptr.From(updated.Location)
	phoneNumber := ptr.From(updated.PhoneNumber)

	ctx.JSON(http.StatusOK, contracts.UpdateAgentResponse{
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

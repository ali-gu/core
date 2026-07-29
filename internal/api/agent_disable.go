package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type DisableAgentHandler struct {
	*API

	uri contracts.AgentURI
}

func (h *DisableAgentHandler) IsWrite() bool {
	return true
}

func (h *DisableAgentHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *DisableAgentHandler) Permissions(ctx *gin.Context, db storage.DB) error {
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

func (h *DisableAgentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	disabled, err := h.Biz.Agent.Disable(ctx, db, h.uri.AgentID)
	if err != nil {
		return rerror.Wrap(err)
	}

	location := ptr.From(disabled.Location)
	phoneNumber := ptr.From(disabled.PhoneNumber)

	ctx.JSON(http.StatusOK, contracts.DisableAgentResponse{
		ID:     disabled.ID,
		Status: disabled.EntityState,
		Name:   disabled.Name,
		Location: &contracts.AgentLocation{
			ID:      ptr.ToPtrOrNil(location.ID),
			Address: ptr.ToPtrOrNil(location.Address),
		},
		PhoneNumber: &contracts.AgentPhoneNumber{
			ID:          ptr.ToPtrOrNil(phoneNumber.ID),
			PhoneNumber: ptr.ToPtrOrNil(phoneNumber.PhoneNumber),
		},
		CreatedAt: disabled.CreatedAt,
	})
	return nil
}

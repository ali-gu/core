package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type CreateTestAgentHandler struct {
	*API

	user storage.User
}

func (h *CreateTestAgentHandler) IsWrite() bool {
	return true
}

func (h *CreateTestAgentHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *CreateTestAgentHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *CreateTestAgentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	created, err := h.Biz.TestAgent.Create(ctx, db, h.user.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	location := ptr.From(created.Location)
	phoneNumber := ptr.From(created.PhoneNumber)

	ctx.JSON(http.StatusCreated, contracts.CreateTestAgentResponse{
		ID:       created.ID,
		Status:   created.EntityState,
		Name:     created.Name,
		AgentRef: created.AgentRef,
		Location: &contracts.AgentLocation{
			ID:      ptr.ToPtrOrNil(location.ID),
			Address: ptr.ToPtrOrNil(location.Address),
		},
		PhoneNumber: &contracts.AgentPhoneNumber{
			ID:          ptr.ToPtrOrNil(phoneNumber.ID),
			PhoneNumber: ptr.ToPtrOrNil(phoneNumber.PhoneNumber),
		},
		CreatedAt: created.CreatedAt,
	})
	return nil
}

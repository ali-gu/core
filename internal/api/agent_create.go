package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type CreateAgentHandler struct {
	*API

	req  contracts.CreateAgentRequest
	user storage.User
}

func (h *CreateAgentHandler) IsWrite() bool {
	return true
}

func (h *CreateAgentHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *CreateAgentHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *CreateAgentHandler) Handle(ctx *gin.Context, db storage.DB) error {
	created, err := h.Biz.Agent.Create(ctx, db, h.user.PracticeID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.CreateAgentResponse{
		ID:        created.ID,
		Name:      created.Name,
		CreatedAt: created.CreatedAt,
	})
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type CreateRoleHandler struct {
	*API

	req contracts.CreateRoleRequest
}

func (h *CreateRoleHandler) IsWrite() bool {
	return true
}

func (h *CreateRoleHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *CreateRoleHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *CreateRoleHandler) Handle(ctx *gin.Context, db storage.DB) error {
	created, err := h.Biz.Role.Create(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.CreateRoleResponse{
		ID:        created.ID,
		Type:      created.Type,
		CreatedAt: created.CreatedAt,
	})
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type CreatePracticeHandler struct {
	*API

	req contracts.CreatePracticeRequest
}

func (h *CreatePracticeHandler) IsWrite() bool {
	return true
}

func (h *CreatePracticeHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *CreatePracticeHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *CreatePracticeHandler) Handle(ctx *gin.Context, db storage.DB) error {
	created, err := h.Biz.Practice.Create(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.CreatePracticeResponse{
		ID:        created.ID,
		Name:      created.Name,
		Email:     created.Email,
		ZipCode:   created.ZipCode,
		Website:   created.Website,
		CreatedAt: created.CreatedAt,
	})
	return nil
}

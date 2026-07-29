package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type UpdatePracticeHandler struct {
	*API

	uri contracts.PracticeURI
	req contracts.UpdatePracticeRequest
}

func (h *UpdatePracticeHandler) IsWrite() bool {
	return true
}

func (h *UpdatePracticeHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	if err := ctx.ShouldBindUri(&h.uri); err != nil {
		return bindErr(err)
	}
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *UpdatePracticeHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	if h.uri.PracticeID != user.PracticeID {
		return rerror.NewMessage("forbidden", rerror.Forbidden)
	}
	return nil
}

func (h *UpdatePracticeHandler) Handle(ctx *gin.Context, db storage.DB) error {
	updated, err := h.Biz.Practice.Update(ctx, db, h.uri.PracticeID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.UpdatePracticeResponse{
		ID:        updated.ID,
		Name:      updated.Name,
		Email:     updated.Email,
		ZipCode:   updated.ZipCode,
		Website:   updated.Website,
		Status:    updated.EntityState,
		CreatedAt: updated.CreatedAt,
	})
	return nil
}

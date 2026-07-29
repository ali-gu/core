package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetPracticeHandler struct {
	*API

	uri contracts.PracticeURI
}

func (h *GetPracticeHandler) IsWrite() bool {
	return false
}

func (h *GetPracticeHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *GetPracticeHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	if h.uri.PracticeID != user.PracticeID {
		return rerror.NewMessage("forbidden", rerror.Forbidden)
	}
	return nil
}

func (h *GetPracticeHandler) Handle(ctx *gin.Context, db storage.DB) error {
	practice, err := h.Biz.Practice.GetByID(ctx, db, h.uri.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.GetPracticeResponse{
		ID:        practice.ID,
		Name:      practice.Name,
		Email:     practice.Email,
		ZipCode:   practice.ZipCode,
		Website:   practice.Website,
		Status:    practice.EntityState,
		CreatedAt: practice.CreatedAt,
	})
	return nil
}

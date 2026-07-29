package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type UpdateLocationHandler struct {
	*API

	uri contracts.LocationURI
	req contracts.UpdateLocationRequest
}

func (h *UpdateLocationHandler) IsWrite() bool {
	return true
}

func (h *UpdateLocationHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	if err := ctx.ShouldBindUri(&h.uri); err != nil {
		return bindErr(err)
	}
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *UpdateLocationHandler) Permissions(ctx *gin.Context, db storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	location, err := h.Biz.Location.GetByID(ctx, db, h.uri.LocationID)
	if err != nil || location.PracticeID != user.PracticeID {
		return rerror.NewMessage("forbidden", rerror.Forbidden)
	}
	return nil
}

func (h *UpdateLocationHandler) Handle(ctx *gin.Context, db storage.DB) error {
	updated, err := h.Biz.Location.Update(ctx, db, h.uri.LocationID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.UpdateLocationResponse{
		ID:        updated.ID,
		Status:    updated.EntityState,
		Address:   updated.Address,
		CreatedAt: updated.CreatedAt,
	})
	return nil
}

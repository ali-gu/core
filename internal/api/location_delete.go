package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type DeleteLocationHandler struct {
	*API

	uri contracts.LocationURI
}

func (h *DeleteLocationHandler) IsWrite() bool {
	return true
}

func (h *DeleteLocationHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *DeleteLocationHandler) Permissions(ctx *gin.Context, db storage.DB) error {
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

func (h *DeleteLocationHandler) Handle(ctx *gin.Context, db storage.DB) error {
	if err := h.Biz.Location.Delete(ctx, db, h.uri.LocationID); err != nil {
		return rerror.Wrap(err)
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

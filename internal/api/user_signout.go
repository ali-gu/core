package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/api/helpers"
	"github.com/ali-gulzar/speechory-core/internal/middleware"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type SignOutHandler struct {
	*API

	accessToken string
}

func (h *SignOutHandler) IsWrite() bool {
	return false
}

func (h *SignOutHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	token, err := middleware.BearerToken(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.accessToken = token
	return nil
}

func (h *SignOutHandler) Permissions(_ *gin.Context, _ storage.DB) error {
	return nil
}

func (h *SignOutHandler) Handle(ctx *gin.Context, db storage.DB) error {
	if err := h.Biz.User.SignOut(ctx, db, h.accessToken); err != nil {
		return rerror.Wrap(err)
	}

	helpers.ClearRefreshCookie(ctx, helpers.RefreshCookieSecure(h.Config.Landscape))
	ctx.Status(http.StatusNoContent)
	return nil
}

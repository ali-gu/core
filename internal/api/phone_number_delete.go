package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type DeletePhoneNumberHandler struct {
	*API

	uri  contracts.PhoneNumberURI
	user storage.User
}

func (h *DeletePhoneNumberHandler) IsWrite() bool {
	return true
}

func (h *DeletePhoneNumberHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *DeletePhoneNumberHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *DeletePhoneNumberHandler) Handle(ctx *gin.Context, db storage.DB) error {
	if err := h.Biz.PhoneNumber.Delete(ctx, db, h.user.PracticeID, h.uri.PhoneNumberID); err != nil {
		return rerror.Wrap(err)
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type DisablePhoneNumberHandler struct {
	*API

	uri  contracts.PhoneNumberURI
	user storage.User
}

func (h *DisablePhoneNumberHandler) IsWrite() bool {
	return true
}

func (h *DisablePhoneNumberHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *DisablePhoneNumberHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *DisablePhoneNumberHandler) Handle(ctx *gin.Context, db storage.DB) error {
	phoneNumber, err := h.Biz.PhoneNumber.Disable(ctx, db, h.user.PracticeID, h.uri.PhoneNumberID)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.DisablePhoneNumberResponse{
		ID:          phoneNumber.ID,
		PhoneNumber: phoneNumber.PhoneNumber,
		EntityState: string(phoneNumber.EntityState),
		DisabledAt:  phoneNumber.DisabledAt,
	})
	return nil
}

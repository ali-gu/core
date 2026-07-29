package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type ReservePhoneNumberHandler struct {
	*API

	req  contracts.ReservePhoneNumberRequest
	user storage.User
}

func (h *ReservePhoneNumberHandler) IsWrite() bool {
	return true
}

func (h *ReservePhoneNumberHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *ReservePhoneNumberHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *ReservePhoneNumberHandler) Handle(ctx *gin.Context, db storage.DB) error {
	pn, err := h.Biz.PhoneNumber.Reserve(ctx, db, h.user.PracticeID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.ReservePhoneNumberResponse{
		ID:          pn.ID,
		PhoneNumber: pn.PhoneNumber,
		EntityState: string(pn.EntityState),
		CreatedAt:   pn.CreatedAt,
	})
	return nil
}

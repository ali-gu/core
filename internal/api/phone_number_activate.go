package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type ActivatePhoneNumberHandler struct {
	*API

	req contracts.ActivatePhoneNumberRequest
}

func (h *ActivatePhoneNumberHandler) IsWrite() bool {
	return true
}

func (h *ActivatePhoneNumberHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *ActivatePhoneNumberHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *ActivatePhoneNumberHandler) Handle(ctx *gin.Context, db storage.DB) error {
	pn, err := h.Biz.PhoneNumber.Activate(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusOK, contracts.ActivatePhoneNumberResponse{
		ID:               pn.ID,
		PhoneNumber:      pn.PhoneNumber,
		PhoneNumberIDRef: pn.PhoneNumberIDRef,
		EntityState:      string(pn.EntityState),
		CreatedAt:        pn.CreatedAt,
		UpdatedAt:        pn.UpdatedAt,
	})
	return nil
}

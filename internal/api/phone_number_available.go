package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type ListAvailablePhoneNumbersHandler struct {
	*API

	req contracts.ListAvailablePhoneNumbersRequest
}

func (h *ListAvailablePhoneNumbersHandler) IsWrite() bool {
	return false
}

func (h *ListAvailablePhoneNumbersHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindQuery(&h.req))
}

func (h *ListAvailablePhoneNumbersHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *ListAvailablePhoneNumbersHandler) Handle(ctx *gin.Context, db storage.DB) error {
	numbers, err := h.Biz.PhoneNumber.ListAvailable(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.AvailablePhoneNumber, len(numbers))
	for i, n := range numbers {
		data[i] = contracts.AvailablePhoneNumber{
			PhoneNumber: n.PhoneNumber,
			Reservable:  n.Reservable,
		}
	}

	ctx.JSON(http.StatusOK, contracts.ListAvailablePhoneNumbersResponse{Data: data})
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type ListPurchasedPhoneNumbersHandler struct {
	*API
}

func (h *ListPurchasedPhoneNumbersHandler) IsWrite() bool {
	return false
}

func (h *ListPurchasedPhoneNumbersHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *ListPurchasedPhoneNumbersHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *ListPurchasedPhoneNumbersHandler) Handle(ctx *gin.Context, db storage.DB) error {
	numbers, err := h.Biz.PhoneNumber.ListPurchased(ctx, db)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.PurchasedPhoneNumber, len(numbers))
	for i, n := range numbers {
		data[i] = contracts.PurchasedPhoneNumber{
			ID:          n.ID,
			PhoneNumber: n.PhoneNumber,
			Status:      n.Status,
		}
	}

	ctx.JSON(http.StatusOK, contracts.ListPurchasedPhoneNumbersResponse{Data: data})
	return nil
}

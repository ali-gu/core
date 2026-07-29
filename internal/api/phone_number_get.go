package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetPhoneNumbersHandler struct {
	*API

	user storage.User
}

func (h *GetPhoneNumbersHandler) IsWrite() bool {
	return false
}

func (h *GetPhoneNumbersHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetPhoneNumbersHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *GetPhoneNumbersHandler) Handle(ctx *gin.Context, db storage.DB) error {
	numbers, err := h.Biz.PhoneNumber.Get(ctx, db, h.user.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.PhoneNumber, len(numbers))
	for i, n := range numbers {
		data[i] = contracts.PhoneNumber{
			ID:          n.ID,
			PhoneNumber: n.PhoneNumber,
			Status:      string(n.EntityState),
			CreatedAt:   n.CreatedAt,
		}
	}

	ctx.JSON(http.StatusOK, contracts.GetPhoneNumbersResponse{Data: data})
	return nil
}

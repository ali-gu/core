package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type CreateLocationHandler struct {
	*API

	req  contracts.CreateLocationRequest
	user storage.User
}

func (h *CreateLocationHandler) IsWrite() bool {
	return true
}

func (h *CreateLocationHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *CreateLocationHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *CreateLocationHandler) Handle(ctx *gin.Context, db storage.DB) error {
	location, ehrRecord, err := h.Biz.Location.Create(ctx, db, h.user.PracticeID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.CreateLocationResponse{
		Location: contracts.Location{
			ID:        location.ID,
			Status:    location.EntityState,
			Address:   location.Address,
			CreatedAt: location.CreatedAt,
		},
		EHR: contracts.EHR{
			OnboardingID:  ehrRecord.OnboardingID,
			OnboardingURL: ehrRecord.OnboardingURL,
		},
	})
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetLocationHandler struct {
	*API

	uri      contracts.LocationURI
	location storage.Location
}

func (h *GetLocationHandler) IsWrite() bool {
	return false
}

func (h *GetLocationHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindUri(&h.uri))
}

func (h *GetLocationHandler) Permissions(ctx *gin.Context, db storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	location, err := h.Biz.Location.GetByID(ctx, db, h.uri.LocationID)
	if err != nil || location.PracticeID != user.PracticeID {
		return rerror.NewMessage("forbidden", rerror.Forbidden)
	}

	h.location = *location
	return nil
}

func (h *GetLocationHandler) Handle(ctx *gin.Context, _ storage.DB) error {
	resp := contracts.Location{
		ID:        h.location.ID,
		Status:    h.location.EntityState,
		Address:   h.location.Address,
		CreatedAt: h.location.CreatedAt,
	}
	if h.location.EHR != nil {
		resp.EHR = &contracts.EHR{
			OnboardingID:  h.location.EHR.OnboardingID,
			OnboardingURL: h.location.EHR.OnboardingURL,
		}
	}

	ctx.JSON(http.StatusOK, resp)
	return nil
}

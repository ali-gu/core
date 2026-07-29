package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetLocationsHandler struct {
	*API

	user storage.User
}

func (h *GetLocationsHandler) IsWrite() bool {
	return false
}

func (h *GetLocationsHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetLocationsHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *GetLocationsHandler) Handle(ctx *gin.Context, db storage.DB) error {
	locations, err := h.Biz.Location.Get(ctx, db, h.user.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.Location, len(locations))
	for i, l := range locations {
		data[i] = contracts.Location{
			ID:        l.ID,
			Status:    l.EntityState,
			Address:   l.Address,
			CreatedAt: l.CreatedAt,
		}
		if l.EHR != nil {
			data[i].EHR = &contracts.EHR{
				OnboardingID:  l.EHR.OnboardingID,
				OnboardingURL: l.EHR.OnboardingURL,
			}
		}
	}

	ctx.JSON(http.StatusOK, contracts.GetLocationsResponse{Data: data})
	return nil
}

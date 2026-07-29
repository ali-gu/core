package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetLocationsAnalyticsHandler struct {
	*API

	user storage.User
}

func (h *GetLocationsAnalyticsHandler) IsWrite() bool {
	return false
}

func (h *GetLocationsAnalyticsHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetLocationsAnalyticsHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *GetLocationsAnalyticsHandler) Handle(ctx *gin.Context, db storage.DB) error {
	locations, err := h.Biz.Location.Get(ctx, db, h.user.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	byState := contracts.EntityStateCounts{}
	data := make([]contracts.Location, len(locations))
	for i, l := range locations {
		byState[string(l.EntityState)]++
		data[i] = contracts.Location{
			ID:        l.ID,
			Status:    l.EntityState,
			Address:   l.Address,
			CreatedAt: l.CreatedAt,
		}
	}

	ctx.JSON(http.StatusOK, contracts.GetLocationsAnalyticsResponse{
		Summary: contracts.LocationsAnalyticsSummary{
			TotalLocations: len(locations),
			ByState:        byState,
		},
		Data: data,
	})
	return nil
}

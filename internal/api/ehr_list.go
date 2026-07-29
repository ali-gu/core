package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetEHRHandler struct {
	*API
}

func (h *GetEHRHandler) IsWrite() bool {
	return false
}

func (h *GetEHRHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetEHRHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *GetEHRHandler) Handle(ctx *gin.Context, _ storage.DB) error {
	ctx.JSON(http.StatusOK, contracts.GetEHRResponse{Data: constants.NexHealthEHRs})
	return nil
}

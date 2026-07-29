package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type CreateToolHandler struct {
	*API
}

func (h *CreateToolHandler) IsWrite() bool {
	return true
}

func (h *CreateToolHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *CreateToolHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	_, err := requireUser(ctx)
	return rerror.Wrap(err)
}

func (h *CreateToolHandler) Handle(ctx *gin.Context, db storage.DB) error {
	tools, err := h.Biz.Tool.Sync(ctx, db)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.Tool, len(tools))
	for i, t := range tools {
		data[i] = toContractTool(t)
	}

	ctx.JSON(http.StatusOK, contracts.GetToolsResponse{Data: data})
	return nil
}

func toContractTool(t storage.Tool) contracts.Tool {
	return contracts.Tool{
		ID:        t.ID,
		Status:    t.EntityState,
		Type:      t.Type,
		Kind:      t.Kind,
		Config:    t.Config,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

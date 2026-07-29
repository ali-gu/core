package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetAgentsHandler struct {
	*API

	user storage.User
}

func (h *GetAgentsHandler) IsWrite() bool {
	return false
}

func (h *GetAgentsHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetAgentsHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *GetAgentsHandler) Handle(ctx *gin.Context, db storage.DB) error {
	agents, err := h.Biz.Agent.Get(ctx, db, h.user.PracticeID)
	if err != nil {
		return rerror.Wrap(err)
	}

	data := make([]contracts.Agent, len(agents))
	for i, a := range agents {
		location := ptr.From(a.Location)
		phoneNumber := ptr.From(a.PhoneNumber)
		item := contracts.Agent{
			ID:     a.ID,
			Status: a.EntityState,
			Location: &contracts.AgentLocation{
				ID:      ptr.ToPtrOrNil(location.ID),
				Address: ptr.ToPtrOrNil(location.Address),
			},
			PhoneNumber: &contracts.AgentPhoneNumber{
				ID:          ptr.ToPtrOrNil(phoneNumber.ID),
				PhoneNumber: ptr.ToPtrOrNil(phoneNumber.PhoneNumber),
			},
			Name:      a.Name,
			CreatedAt: a.CreatedAt,
		}
		data[i] = item
	}

	ctx.JSON(http.StatusOK, contracts.GetAgentsResponse{Data: data})
	return nil
}

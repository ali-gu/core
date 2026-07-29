package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type GetUserHandler struct {
	*API

	user storage.User
}

func (h *GetUserHandler) IsWrite() bool {
	return false
}

func (h *GetUserHandler) Setup(_ *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return nil
}

func (h *GetUserHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *GetUserHandler) Handle(ctx *gin.Context, _ storage.DB) error {
	ctx.JSON(http.StatusOK, contracts.GetUserResponse{
		ID:          h.user.ID,
		Email:       h.user.Email,
		RoleID:      h.user.RoleID,
		PracticeID:  h.user.PracticeID,
		EntityState: h.user.EntityState,
		CreatedAt:   h.user.CreatedAt,
		Practice: contracts.UserPractice{
			ID:     h.user.Practice.ID,
			Name:   h.user.Practice.Name,
			Status: h.user.Practice.EntityState,
		},
	})
	return nil
}

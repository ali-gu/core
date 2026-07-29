package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type InviteUserHandler struct {
	*API

	req  contracts.InviteUserRequest
	user storage.User
}

func (h *InviteUserHandler) IsWrite() bool {
	return true
}

func (h *InviteUserHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *InviteUserHandler) Permissions(ctx *gin.Context, _ storage.DB) error {
	user, err := requireUser(ctx)
	if err != nil {
		return rerror.Wrap(err)
	}

	h.user = user
	return nil
}

func (h *InviteUserHandler) Handle(ctx *gin.Context, db storage.DB) error {
	invited, err := h.Biz.User.Invite(ctx, db, h.user.PracticeID, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.InviteUserResponse{
		ID:          invited.ID,
		Email:       invited.Email,
		RoleID:      invited.RoleID,
		PracticeID:  invited.PracticeID,
		EntityState: invited.EntityState,
		CreatedAt:   invited.CreatedAt,
	})
	return nil
}

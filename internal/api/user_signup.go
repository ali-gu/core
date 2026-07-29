package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type SignUpHandler struct {
	*API

	req contracts.SignUpRequest
}

func (h *SignUpHandler) IsWrite() bool {
	return true
}

func (h *SignUpHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *SignUpHandler) Permissions(_ *gin.Context, _ storage.DB) error {
	return nil
}

func (h *SignUpHandler) Handle(ctx *gin.Context, db storage.DB) error {
	created, err := h.Biz.User.SignUp(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	ctx.JSON(http.StatusCreated, contracts.SignUpResponse{
		ID:          created.ID,
		Email:       created.Email,
		RoleID:      created.RoleID,
		PracticeID:  created.PracticeID,
		EntityState: created.EntityState,
		CreatedAt:   created.CreatedAt,
	})
	return nil
}

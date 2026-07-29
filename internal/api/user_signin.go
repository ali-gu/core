package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/api/helpers"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type SignInHandler struct {
	*API

	req contracts.SignInRequest
}

func (h *SignInHandler) IsWrite() bool {
	return false
}

func (h *SignInHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api
	return bindErr(ctx.ShouldBindJSON(&h.req))
}

func (h *SignInHandler) Permissions(_ *gin.Context, _ storage.DB) error {
	return nil
}

func (h *SignInHandler) Handle(ctx *gin.Context, db storage.DB) error {
	user, session, err := h.Biz.User.SignIn(ctx, db, h.req)
	if err != nil {
		return rerror.Wrap(err)
	}

	helpers.SetRefreshCookie(ctx, session.RefreshToken, helpers.RefreshCookieSecure(h.Config.Landscape))

	ctx.JSON(http.StatusOK, contracts.SignInResponse{
		GetUserResponse: contracts.GetUserResponse{
			ID:          user.ID,
			Email:       user.Email,
			RoleID:      user.RoleID,
			PracticeID:  user.PracticeID,
			EntityState: user.EntityState,
			CreatedAt:   user.CreatedAt,
			Practice: contracts.UserPractice{
				ID:     user.Practice.ID,
				Name:   user.Practice.Name,
				Status: user.Practice.EntityState,
			},
		},
		AccessToken: session.AccessToken,
		TokenType:   session.TokenType,
		ExpiresIn:   session.ExpiresIn,
		ExpiresAt:   session.ExpiresAt,
	})
	return nil
}

package api

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/api/helpers"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

type RefreshHandler struct {
	*API

	refreshToken string
}

func (h *RefreshHandler) IsWrite() bool {
	return false
}

func (h *RefreshHandler) Setup(ctx *gin.Context, _ storage.DB, api *API) error {
	h.API = api

	token, err := ctx.Cookie(helpers.RefreshCookieName)
	if err != nil || token == "" {
		helpers.ClearRefreshCookie(ctx, helpers.RefreshCookieSecure(h.Config.Landscape))
		return rerror.NewMessage("missing refresh token", rerror.Permission)
	}

	h.refreshToken = token
	return nil
}

func (h *RefreshHandler) Permissions(_ *gin.Context, _ storage.DB) error {
	return nil
}

func (h *RefreshHandler) Handle(ctx *gin.Context, db storage.DB) error {
	user, session, err := h.Biz.User.Refresh(ctx, db, h.refreshToken)
	if err != nil {
		helpers.ClearRefreshCookie(ctx, helpers.RefreshCookieSecure(h.Config.Landscape))
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

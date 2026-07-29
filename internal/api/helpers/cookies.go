package helpers

import (
	"net/http"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/gin-gonic/gin"
)

const (
	RefreshCookieName = "refresh_token"
	RefreshCookiePath = "/v1/users"
)

var refreshCookieMaxAge = int((30 * 24 * time.Hour).Seconds())

func RefreshCookieSecure(landscape constants.Landscape) bool {
	return landscape != constants.LandscapeLocal
}

func SetRefreshCookie(ctx *gin.Context, token string, secure bool) {
	ctx.SetCookieData(&http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     RefreshCookiePath,
		MaxAge:   refreshCookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearRefreshCookie(ctx *gin.Context, secure bool) {
	ctx.SetCookieData(&http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

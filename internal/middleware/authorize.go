package middleware

import (
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"

	"github.com/gin-gonic/gin"
)

func RequireRole(role constants.RoleType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, ok := rctx.GetAuthenticatedUser(ctx)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if user.Role.Type != role {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		ctx.Next()
	}
}

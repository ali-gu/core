package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var AllowedOrigins = map[string]bool{
	"http://localhost:5173":              true,
	"https://dev.platform.speechory.com": true,
	"https://platform.speechory.com":     true,
}

func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if !AllowedOrigins[origin] {
			if ctx.Request.Method == http.MethodOptions {
				ctx.AbortWithStatus(http.StatusNoContent)
				return
			}
			ctx.Next()
			return
		}

		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Vary", "Origin")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

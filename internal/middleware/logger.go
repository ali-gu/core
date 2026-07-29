package middleware

import (
	"github.com/ali-gulzar/speechory-core/pkg/rlog"

	"github.com/gin-gonic/gin"
)

func Logger(log rlog.ILogger) gin.HandlerFunc {
	if log == nil {
		log = rlog.Nop()
	}

	return func(ctx *gin.Context) {
		path := ctx.FullPath()
		if path == "" {
			path = ctx.Request.URL.Path
		}

		requestLogger := log.
			AddMetadata("method", ctx.Request.Method).
			AddMetadata("path", path)
		requestLogger.Infof("%s %s", ctx.Request.Method, path)

		rlog.WithLogger(ctx, requestLogger)
		ctx.Next()
	}
}

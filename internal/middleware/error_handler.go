package middleware

import (
	"errors"
	"net/http"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/pkg/rlog"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := rlog.GetLogger(ctx)

		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors[0].Err

		var respErr contracts.ResponseError
		if errors.As(err, &respErr) {
			ctx.JSON(respErr.StatusCode, gin.H{"error": respErr.Error()})
			return
		}

		var rErr *rerror.Error
		if errors.As(err, &rErr) {
			logger.Debugf("%s", rErr)
			respErr = contracts.ResponseErrorFromError(*rErr)
			ctx.JSON(respErr.StatusCode, gin.H{"error": respErr.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

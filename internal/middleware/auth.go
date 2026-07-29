package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/pkg/rlog"
	"github.com/gin-gonic/gin"
)

type Authenticator interface {
	Authenticate(ctx context.Context, db storage.DB, accessToken string) (*storage.User, error)
}

func Auth(dbMux storage.IDBMux, authenticator Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := BearerToken(ctx)
		if err != nil {
			abortWithError(ctx, err)
			return
		}

		db, err := dbMux.BeginRead()
		if err != nil {
			abortWithError(ctx, rerror.New(err))
			return
		}

		user, err := authenticator.Authenticate(ctx, db, token)
		if err != nil {
			abortWithError(ctx, err)
			return
		}

		rctx.WithAuthenticatedUser(ctx, *user)
		ctx.Next()
	}
}

func BearerToken(ctx *gin.Context) (string, error) {
	header := ctx.GetHeader("Authorization")
	if header == "" {
		return "", rerror.NewMessage("missing authorization header", rerror.Permission)
	}

	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", rerror.NewMessage("malformed authorization header", rerror.Permission)
	}

	return strings.TrimSpace(token), nil
}

func abortWithError(ctx *gin.Context, err error) {
	var rerr *rerror.Error
	if !errors.As(err, &rerr) {
		rerr = rerror.New(err)
	}
	rlog.GetLogger(ctx).Debugf("%s", rerr.Error())
	respErr := contracts.ResponseErrorFromError(*rerr)
	ctx.AbortWithStatusJSON(respErr.StatusCode, gin.H{"error": respErr.Error()})
}

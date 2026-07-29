package api

import (
	"bytes"
	"errors"
	"io"

	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"

	"github.com/gin-gonic/gin"
)

var ErrInvalidSignature = errors.New("invalid webhook signature")

type WebhookHandler interface {
	Validator(ctx *gin.Context, api *API, body []byte) error
	Handle(ctx *gin.Context, db storage.DB) error
}

func (api *API) BuildWebhook(factory func() WebhookHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := factory()

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			_ = ctx.Error(asRerror(err, rerror.Validation))
			ctx.Abort()
			return
		}
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

		if err = h.Validator(ctx, api, body); err != nil {
			fallback := rerror.Validation
			if errors.Is(err, ErrInvalidSignature) {
				fallback = rerror.Permission
			}
			_ = ctx.Error(asRerror(err, fallback))
			ctx.Abort()
			return
		}

		db, err := api.DBMux.BeginRead()
		if err != nil {
			_ = ctx.Error(asRerror(err, rerror.Internal))
			ctx.Abort()
			return
		}

		if err = h.Handle(ctx, db); err != nil {
			_ = ctx.Error(asRerror(err, rerror.Internal))
			ctx.Abort()
			return
		}
	}
}

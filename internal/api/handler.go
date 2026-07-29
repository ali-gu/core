package api

import (
	"errors"
	"io"

	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
)

// asRerror normalizes any error into an *rerror.Error. Handlers are expected to
// return rerrors already (via rerror.Wrap / rerror.NewMessage / bindErr); this
// is a safety net that also assigns the per-phase fallback kind to any stray
// non-rerror error.
func asRerror(err error, fallback rerror.Kind) *rerror.Error {
	var rerr *rerror.Error
	if errors.As(err, &rerr) {
		return rerr
	}
	return rerror.New(err).WithKind(fallback)
}

// bindErr wraps a gin request-binding result as a Validation rerror so a
// malformed request surfaces as a 400 with a client-facing message. nil stays
// nil, so it is safe to use directly on a Setup return.
func bindErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return rerror.NewMessage("request body is required", rerror.Validation)
	}
	return rerror.New(err).WithKind(rerror.Validation)
}

func requireUser(ctx *gin.Context) (storage.User, error) {
	user, ok := rctx.GetAuthenticatedUser(ctx)
	if !ok {
		return storage.User{}, rerror.NewMessage("authentication required", rerror.Permission)
	}
	return user, nil
}

type Handler interface {
	IsWrite() bool
	Setup(ctx *gin.Context, db storage.DB, api *API) error
	Permissions(ctx *gin.Context, db storage.DB) error
	Handle(ctx *gin.Context, db storage.DB) error
}

func (api *API) Build(factory func() Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := factory()

		var db storage.DB
		var err error
		if h.IsWrite() {
			db, err = api.DBMux.BeginWrite()
		} else {
			db, err = api.DBMux.BeginRead()
		}
		if err != nil {
			_ = ctx.Error(asRerror(err, rerror.Internal))
			ctx.Abort()
			return
		}

		if err = h.Setup(ctx, db, api); err != nil {
			_ = ctx.Error(asRerror(err, rerror.Validation))
			ctx.Abort()
			return
		}

		if err = h.Permissions(ctx, db); err != nil {
			_ = ctx.Error(asRerror(err, rerror.Permission))
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

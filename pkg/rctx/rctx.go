package rctx

import (
	"context"

	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	contextKeyAuthenticatedUser ContextKey = "speechory_authenticated_user"
	ContextKeyLogger            ContextKey = "speechory_logger"
)

func Set[T context.Context](ctx T, key ContextKey, value any) T {
	switch ctxValue := any(ctx).(type) {
	case *gin.Context:
		ctxValue.Set(string(key), value)
	case context.Context:
		ctx = context.WithValue(ctx, string(key), value).(T)
	}
	return ctx
}

func Get[T context.Context, K any](ctx T, key ContextKey) (K, bool) {
	var noop K
	switch ctxValue := any(ctx).(type) {
	case *gin.Context:
		v, ok := ctxValue.Get(string(key))
		if !ok {
			return noop, false
		} else if v, ok := v.(K); ok {
			return v, true
		}
	case context.Context:
		v, ok := ctxValue.Value(string(key)).(K)
		return v, ok
	}
	return noop, false
}

func WithAuthenticatedUser[T context.Context](ctx T, user storage.User) {
	Set(ctx, contextKeyAuthenticatedUser, user)
}

func GetAuthenticatedUser[T context.Context](ctx T) (user storage.User, exists bool) {
	return Get[T, storage.User](ctx, contextKeyAuthenticatedUser)
}

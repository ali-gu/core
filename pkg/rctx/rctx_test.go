package rctx_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testKey rctx.ContextKey = "rctx_test_key"

func newGinContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}

func Test_Set_Get(t *testing.T) {
	t.Run("gin_context_round_trips_a_value", func(t *testing.T) {
		ctx := newGinContext()

		rctx.Set(ctx, testKey, "hello")

		got, ok := rctx.Get[*gin.Context, string](ctx, testKey)
		require.True(t, ok)
		require.Equal(t, "hello", got)
	})

	t.Run("gin_context_missing_key_returns_zero_value", func(t *testing.T) {
		ctx := newGinContext()

		got, ok := rctx.Get[*gin.Context, string](ctx, testKey)
		require.False(t, ok)
		require.Equal(t, "", got)
	})

	t.Run("gin_context_wrong_type_returns_zero_value", func(t *testing.T) {
		ctx := newGinContext()
		rctx.Set(ctx, testKey, "hello")

		got, ok := rctx.Get[*gin.Context, int](ctx, testKey)
		require.False(t, ok)
		require.Equal(t, 0, got)
	})

	t.Run("plain_context_round_trips_a_value", func(t *testing.T) {
		ctx := rctx.Set[context.Context](context.Background(), testKey, "hello")

		got, ok := rctx.Get[context.Context, string](ctx, testKey)
		require.True(t, ok)
		require.Equal(t, "hello", got)
	})

	t.Run("plain_context_missing_key_returns_zero_value", func(t *testing.T) {
		got, ok := rctx.Get[context.Context, string](context.Background(), testKey)
		require.False(t, ok)
		require.Equal(t, "", got)
	})

	t.Run("plain_context_wrong_type_returns_zero_value", func(t *testing.T) {
		ctx := rctx.Set[context.Context](context.Background(), testKey, "hello")

		got, ok := rctx.Get[context.Context, int](ctx, testKey)
		require.False(t, ok)
		require.Equal(t, 0, got)
	})
}

func Test_WithAuthenticatedUser_GetAuthenticatedUser(t *testing.T) {
	t.Run("gin_context_round_trips_the_user", func(t *testing.T) {
		ctx := newGinContext()
		user := storage.User{ID: ksuid.New(), PracticeID: ksuid.New()}

		rctx.WithAuthenticatedUser(ctx, user)

		got, ok := rctx.GetAuthenticatedUser(ctx)
		require.True(t, ok)
		require.Equal(t, user.ID, got.ID)
		require.Equal(t, user.PracticeID, got.PracticeID)
	})

	t.Run("gin_context_missing_user_returns_not_found", func(t *testing.T) {
		ctx := newGinContext()

		_, ok := rctx.GetAuthenticatedUser(ctx)
		require.False(t, ok)
	})
}

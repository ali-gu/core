package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/middleware"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_RequireRole(t *testing.T) {
	t.Run("continues_when_the_user_has_the_required_role", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		rctx.WithAuthenticatedUser(c, storage.User{Role: storage.Role{Type: constants.RoleTypeSuperAdmin}})

		middleware.RequireRole(constants.RoleTypeSuperAdmin)(c)

		require.False(t, c.IsAborted())
	})

	t.Run("aborts_403_when_the_user_has_a_different_role", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		rctx.WithAuthenticatedUser(c, storage.User{Role: storage.Role{Type: constants.RoleTypeAdmin}})

		middleware.RequireRole(constants.RoleTypeSuperAdmin)(c)

		require.True(t, c.IsAborted())
		require.Equal(t, http.StatusForbidden, c.Writer.Status())
	})

	t.Run("aborts_401_when_there_is_no_authenticated_user", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		middleware.RequireRole(constants.RoleTypeSuperAdmin)(c)

		require.True(t, c.IsAborted())
		require.Equal(t, http.StatusUnauthorized, c.Writer.Status())
	})
}

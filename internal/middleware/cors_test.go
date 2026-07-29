package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_CORS(t *testing.T) {
	allowed := []string{
		"http://localhost:5173",
		"https://dev.platform.speechory.com",
		"https://platform.speechory.com",
	}

	for _, origin := range allowed {
		t.Run("adds_cors_headers_for_"+origin, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/practices/123", nil)
			c.Request.Header.Set("Origin", origin)

			middleware.CORS()(c)

			require.False(t, c.IsAborted())
			require.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
			require.Equal(t, "Origin", w.Header().Get("Vary"))
			require.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		})
	}

	t.Run("does_not_add_cors_headers_for_a_disallowed_origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/practices/123", nil)
		c.Request.Header.Set("Origin", "http://evil.example.com")

		middleware.CORS()(c)

		require.False(t, c.IsAborted())
		require.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		require.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("responds_to_preflight_for_allowed_origin", func(t *testing.T) {
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(middleware.CORS())
		called := false
		router.OPTIONS("/practices/:practice_id", func(c *gin.Context) { called = true })

		req := httptest.NewRequest(http.MethodOptions, "/practices/123", nil)
		req.Header.Set("Origin", "https://platform.speechory.com")
		router.ServeHTTP(w, req)

		require.False(t, called)
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Equal(t, "https://platform.speechory.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("aborts_preflight_for_disallowed_origin", func(t *testing.T) {
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(middleware.CORS())
		called := false
		router.OPTIONS("/practices/:practice_id", func(c *gin.Context) { called = true })

		req := httptest.NewRequest(http.MethodOptions, "/practices/123", nil)
		req.Header.Set("Origin", "http://evil.example.com")
		router.ServeHTTP(w, req)

		require.False(t, called)
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

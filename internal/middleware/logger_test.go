package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/middleware"
	"github.com/ali-gulzar/speechory-core/pkg/rlog"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeLogger struct {
	infof    []string
	metadata map[string]any
}

func (f *fakeLogger) Debug(string)          {}
func (f *fakeLogger) Debugf(string, ...any) {}
func (f *fakeLogger) Info(string)           {}
func (f *fakeLogger) Infof(format string, args ...any) {
	f.infof = append(f.infof, fmt.Sprintf(format, args...))
}
func (f *fakeLogger) Warn(string)           {}
func (f *fakeLogger) Warnf(string, ...any)  {}
func (f *fakeLogger) Error(string)          {}
func (f *fakeLogger) Errorf(string, ...any) {}

func (f *fakeLogger) AddMetadata(key string, value any) rlog.ILogger {
	if f.metadata == nil {
		f.metadata = map[string]any{}
	}
	f.metadata[key] = value
	return f
}

func Test_Logger(t *testing.T) {
	t.Run("logs_and_attaches_the_logger_when_no_route_matched", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/practices/123", nil)

		fake := &fakeLogger{}
		middleware.Logger(fake)(c)

		require.False(t, c.IsAborted())
		require.Equal(t, []string{"GET /practices/123"}, fake.infof)
		require.Equal(t, "GET", fake.metadata["method"])
		require.Equal(t, "/practices/123", fake.metadata["path"])
		require.Equal(t, rlog.ILogger(fake), rlog.GetLogger(c))
	})

	t.Run("uses_the_matched_route_pattern_as_the_path", func(t *testing.T) {
		fake := &fakeLogger{}

		router := gin.New()
		router.Use(middleware.Logger(fake))
		router.GET("/practices/:practice_id", func(c *gin.Context) {})

		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/practices/123", nil))

		require.Equal(t, []string{"GET /practices/:practice_id"}, fake.infof)
		require.Equal(t, "/practices/:practice_id", fake.metadata["path"])
	})

	t.Run("a_nil_logger_does_not_panic", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		require.NotPanics(t, func() { middleware.Logger(nil)(c) })
		require.False(t, c.IsAborted())
	})
}

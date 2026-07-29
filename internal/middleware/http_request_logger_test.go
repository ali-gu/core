package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/middleware"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rctx"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

type fakeHTTPReqStorage struct {
	created storage.HTTPRequest
}

func (f *fakeHTTPReqStorage) Create(_ context.Context, _ storage.DB, req storage.HTTPRequest) error {
	f.created = req
	return nil
}

func (f *fakeHTTPReqStorage) Get(_ context.Context, _ storage.DB, _ storage.HTTPRequestFilters) ([]storage.HTTPRequest, error) {
	return nil, nil
}

type failingHTTPReqStorage struct{}

func (f *failingHTTPReqStorage) Create(_ context.Context, _ storage.DB, _ storage.HTTPRequest) error {
	return rerror.NewMessage("db write failed", rerror.Internal)
}

func (f *failingHTTPReqStorage) Get(_ context.Context, _ storage.DB, _ storage.HTTPRequestFilters) ([]storage.HTTPRequest, error) {
	return nil, nil
}

type fakeWriteDBMux struct {
	storage.IDBMux
}

func (fakeWriteDBMux) BeginWrite() (storage.DB, error) { return nil, nil }

func Test_HTTPRequestLogger(t *testing.T) {
	t.Run("captures_request_with_authenticated_user", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/agents", nil)

		user := storage.User{ID: ksuid.New(), PracticeID: ksuid.New()}
		rctx.WithAuthenticatedUser(c, user)

		mw(c)

		require.False(t, c.IsAborted())
		require.Equal(t, "GET", fakeStorage.created.Method)
		require.Equal(t, "/v1/agents", fakeStorage.created.Path)
		require.NotEmpty(t, fakeStorage.created.IPAddress)
		require.NotNil(t, fakeStorage.created.UserID)
		require.Equal(t, user.ID, *fakeStorage.created.UserID)
		require.NotNil(t, fakeStorage.created.PracticeID)
		require.Equal(t, user.PracticeID, *fakeStorage.created.PracticeID)
	})

	t.Run("captures_request_without_authentication", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

		mw(c)

		require.False(t, c.IsAborted())
		require.Equal(t, "GET", fakeStorage.created.Method)
		require.Equal(t, "/health", fakeStorage.created.Path)
		require.NotEmpty(t, fakeStorage.created.IPAddress)
		require.Nil(t, fakeStorage.created.UserID)
		require.Nil(t, fakeStorage.created.PracticeID)
	})

	t.Run("captures_request_body", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString(`{"name":"test agent"}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/agents", body)
		c.Request.Header.Set("Content-Type", "application/json")

		mw(c)

		require.Equal(t, `{"name":"test agent"}`, fakeStorage.created.RequestBody)
	})

	t.Run("captures_response_body_and_status_code", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/agents", nil)

		c.Set("_handler", true)
		mw(c)

		require.Equal(t, http.StatusOK, fakeStorage.created.StatusCode)
	})

	t.Run("captures_status_code_from_handler", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		router := gin.New()
		router.Use(mw)
		router.POST("/v1/agents", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "123"})
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/agents", nil))

		require.Equal(t, http.StatusCreated, fakeStorage.created.StatusCode)
		require.Equal(t, "POST", fakeStorage.created.Method)
		require.Contains(t, fakeStorage.created.ResponseBody, "123")
	})

	t.Run("uses_matched_route_pattern_as_path", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		router := gin.New()
		router.Use(mw)
		router.GET("/practices/:practice_id", func(c *gin.Context) {})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/practices/abc123", nil))

		require.Equal(t, "GET", fakeStorage.created.Method)
		require.Equal(t, "/practices/:practice_id", fakeStorage.created.Path)
	})

	t.Run("handles_empty_request_body", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

		mw(c)

		require.Empty(t, fakeStorage.created.RequestBody)
	})

	t.Run("does_not_abort_the_request", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		mw(c)

		require.False(t, c.IsAborted())
	})

	t.Run("logs_error_when_db_write_fails", func(t *testing.T) {
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, &failingHTTPReqStorage{})

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		require.NotPanics(t, func() { mw(c) })
		require.False(t, c.IsAborted())
	})

	t.Run("masks_password_in_request_body", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString(`{"email":"local@speechory.com","password":"random123","practice_name":"Dentistry 101"}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/users", body)

		mw(c)

		require.Equal(t, `{"email":"local@speechory.com","password":"***","practice_name":"Dentistry 101"}`, fakeStorage.created.RequestBody)
	})

	t.Run("masks_multiple_sensitive_fields", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString(`{"api_key":"abc123","secret":"s3cr3t","data":"safe"}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/webhooks", body)

		mw(c)

		require.Equal(t, `{"api_key":"***","data":"safe","secret":"***"}`, fakeStorage.created.RequestBody)
	})

	t.Run("masks_sensitive_fields_in_nested_objects", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString(`{"user":{"name":"Alice","password":"pass123"},"token":"tok_abc"}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/users", body)

		mw(c)

		require.Equal(t, `{"token":"***","user":{"name":"Alice","password":"***"}}`, fakeStorage.created.RequestBody)
	})

	t.Run("masks_sensitive_fields_in_arrays", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString(`{"items":[{"password":"a","name":"b"},{"secret":"c"}]}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/items", body)

		mw(c)

		require.Equal(t, `{"items":[{"name":"b","password":"***"},{"secret":"***"}]}`, fakeStorage.created.RequestBody)
	})

	t.Run("compacts_pretty_json_request_body", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString("{\n  \"email\": \"test@example.com\",\n  \"password\": \"secret\"\n}")
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/users", body)

		mw(c)

		require.Equal(t, `{"email":"test@example.com","password":"***"}`, fakeStorage.created.RequestBody)
	})

	t.Run("compacts_and_masks_response_body", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		router := gin.New()
		router.Use(mw)
		router.POST("/v1/users", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"id":       "123",
				"email":    "test@example.com",
				"password": "should_not_appear",
			})
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/users", nil))

		require.Contains(t, fakeStorage.created.ResponseBody, `"password":"***"`)
		require.NotContains(t, fakeStorage.created.ResponseBody, "should_not_appear")
	})

	t.Run("leaves_non_json_body_untouched", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString("plain text body")
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/upload", body)

		mw(c)

		require.Equal(t, "plain text body", fakeStorage.created.RequestBody)
	})

	t.Run("case_insensitive_field_matching", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		body := bytes.NewBufferString(`{"Password":"secret","TOKEN":"abc","Safe":"ok"}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/users", body)

		mw(c)

		require.Equal(t, `{"Password":"***","Safe":"ok","TOKEN":"***"}`, fakeStorage.created.RequestBody)
	})

	t.Run("captures_query_params", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/agents?page=2&limit=10", nil)

		mw(c)

		require.NotEmpty(t, fakeStorage.created.QueryParams)
		var parsed map[string][]string
		require.NoError(t, json.Unmarshal([]byte(fakeStorage.created.QueryParams), &parsed))
		require.Equal(t, []string{"2"}, parsed["page"])
		require.Equal(t, []string{"10"}, parsed["limit"])
	})

	t.Run("empty_query_params_when_none_present", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/agents", nil)

		mw(c)

		require.Empty(t, fakeStorage.created.QueryParams)
	})

	t.Run("captures_headers", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-Request-Id", "req-123")

		mw(c)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(fakeStorage.created.Headers), &parsed))
		require.Equal(t, "application/json", parsed["Content-Type"])
		require.Equal(t, "req-123", parsed["X-Request-Id"])
	})

	t.Run("masks_authorization_header", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
		c.Request.Header.Set("Authorization", "Bearer secret_token_123")
		c.Request.Header.Set("Content-Type", "application/json")

		mw(c)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(fakeStorage.created.Headers), &parsed))
		require.Equal(t, "***", parsed["Authorization"])
		require.Equal(t, "application/json", parsed["Content-Type"])
	})

	t.Run("masks_cookie_header", func(t *testing.T) {
		fakeStorage := &fakeHTTPReqStorage{}
		mw := middleware.HTTPRequestLogger(fakeWriteDBMux{}, fakeStorage)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
		c.Request.Header.Set("Cookie", "session=abc123")

		mw(c)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(fakeStorage.created.Headers), &parsed))
		require.Equal(t, "***", parsed["Cookie"])
	})
}

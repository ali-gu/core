package middleware_test

import (
	"context"
	"errors"
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

func init() {
	gin.SetMode(gin.TestMode)
}

type fakeDBMux struct {
	storage.IDBMux
}

func (fakeDBMux) BeginRead() (storage.DB, error) { return nil, nil }

type fakeAuthenticator struct {
	user *storage.User
	err  error
}

func (f fakeAuthenticator) Authenticate(_ context.Context, _ storage.DB, _ string) (*storage.User, error) {
	return f.user, f.err
}

func requestWithAuth(header string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		c.Request.Header.Set("Authorization", header)
	}
	return c
}

func Test_Auth(t *testing.T) {
	t.Run("success_attaches_the_user_and_continues", func(t *testing.T) {
		user := storage.User{ID: ksuid.New(), PracticeID: ksuid.New()}
		mw := middleware.Auth(fakeDBMux{}, fakeAuthenticator{user: &user})

		c := requestWithAuth("Bearer token")
		mw(c)

		require.False(t, c.IsAborted())
		got, ok := rctx.GetAuthenticatedUser(c)
		require.True(t, ok)
		require.Equal(t, user.ID, got.ID)
		require.Equal(t, user.PracticeID, got.PracticeID)
	})

	t.Run("aborts_401_when_the_authorization_header_is_missing", func(t *testing.T) {
		mw := middleware.Auth(fakeDBMux{}, fakeAuthenticator{})

		c := requestWithAuth("")
		mw(c)

		require.True(t, c.IsAborted())
		require.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		_, ok := rctx.GetAuthenticatedUser(c)
		require.False(t, ok)
	})

	t.Run("aborts_401_when_the_token_is_rejected", func(t *testing.T) {
		mw := middleware.Auth(fakeDBMux{}, fakeAuthenticator{err: rerror.NewMessage("invalid token", rerror.Permission)})

		c := requestWithAuth("Bearer token")
		mw(c)

		require.True(t, c.IsAborted())
		require.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		_, ok := rctx.GetAuthenticatedUser(c)
		require.False(t, ok)
	})

	t.Run("aborts_500_when_authentication_fails_internally", func(t *testing.T) {
		mw := middleware.Auth(fakeDBMux{}, fakeAuthenticator{err: errors.New("db down")})

		c := requestWithAuth("Bearer token")
		mw(c)

		require.True(t, c.IsAborted())
		require.Equal(t, http.StatusInternalServerError, c.Writer.Status())
		_, ok := rctx.GetAuthenticatedUser(c)
		require.False(t, ok)
	})
}

func Test_bearerToken(t *testing.T) {
	cases := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "valid", header: "Bearer abc.123", want: "abc.123"},
		{name: "scheme_is_case_insensitive", header: "bearer abc.123", want: "abc.123"},
		{name: "trims_surrounding_space", header: "Bearer   abc.123  ", want: "abc.123"},
		{name: "missing_header", header: "", wantErr: true},
		{name: "no_scheme", header: "abc.123", wantErr: true},
		{name: "wrong_scheme", header: "Basic abc.123", wantErr: true},
		{name: "empty_token", header: "Bearer ", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := middleware.BearerToken(requestWithAuth(tc.header))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

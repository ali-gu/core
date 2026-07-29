package api_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Refresh(t *testing.T) {
	t.Run("200_rotates_the_refresh_cookie_and_returns_a_fresh_access_token", func(t *testing.T) {
		env := newE2E(t)
		practice := fixtures.NewPractice(t, env.cfg, env.bz)
		role := env.seedRole(constants.RoleTypeAdmin)

		const ref = "ref_refresh"
		require.NoError(t, env.cfg.Deps.Storage.User.Create(env.cfg.Ctx, env.cfg.DB, storage.User{
			EntityBase: storage.EntityBase[states.UserState]{EntityState: states.UserStateActive},
			ID:         ksuid.New(),
			UserRef:    ref,
			RoleID:     role.ID,
			PracticeID: practice.ID,
			Email:      "refresh@example.com",
			CreatedAt:  time.Now(),
		}))
		env.authMock.On("RefreshToken", mock.Anything, "old-refresh-token").
			Return(&auth.SignInResult{
				AuthenticatedUser: auth.AuthenticatedUser{ID: ref, Email: "refresh@example.com"},
				Session: auth.Session{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
					TokenType:    "bearer",
					ExpiresIn:    3600,
					ExpiresAt:    1234567890,
				},
			}, nil).Once()

		w := env.doWithCookie(http.MethodPost, "/v1/users/refresh", "", &http.Cookie{
			Name:  "refresh_token",
			Value: "old-refresh-token",
		}, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.SignInResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "refresh@example.com", resp.Email)
		require.Equal(t, practice.ID, resp.PracticeID)
		require.Equal(t, "new-access-token", resp.AccessToken)
		require.Equal(t, "bearer", resp.TokenType)
		require.NotContains(t, w.Body.String(), "new-refresh-token")

		cookie := responseCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		require.Equal(t, "new-refresh-token", cookie.Value)
		require.True(t, cookie.HttpOnly)
		require.True(t, cookie.Secure)
		require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
		require.Equal(t, "/v1/users", cookie.Path)
		require.Greater(t, cookie.MaxAge, 0)
	})

	t.Run("401_and_clears_the_cookie_when_no_refresh_cookie_is_present", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWithCookie(http.MethodPost, "/v1/users/refresh", "", nil, nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)

		cookie := responseCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		require.Empty(t, cookie.Value)
		require.Less(t, cookie.MaxAge, 0)
	})

	t.Run("401_and_clears_the_cookie_when_the_refresh_token_is_invalid", func(t *testing.T) {
		env := newE2E(t)
		env.authMock.On("RefreshToken", mock.Anything, "bad-refresh-token").
			Return(nil, rerror.New(errors.New("invalid refresh token")).WithKind(rerror.Permission)).Once()

		w := env.doWithCookie(http.MethodPost, "/v1/users/refresh", "", &http.Cookie{
			Name:  "refresh_token",
			Value: "bad-refresh-token",
		}, nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)

		cookie := responseCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		require.Empty(t, cookie.Value)
		require.Less(t, cookie.MaxAge, 0)
	})
}

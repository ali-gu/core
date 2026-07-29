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

func Test_SignIn(t *testing.T) {
	t.Run("200_signs_in_and_returns_the_user_with_their_practice_and_session", func(t *testing.T) {
		env := newE2E(t)
		practice := fixtures.NewPractice(t, env.cfg, env.bz)
		role := env.seedRole(constants.RoleTypeAdmin)

		const ref = "ref_signin"
		require.NoError(t, env.cfg.Deps.Storage.User.Create(env.cfg.Ctx, env.cfg.DB, storage.User{
			EntityBase: storage.EntityBase[states.UserState]{EntityState: states.UserStateActive},
			ID:         ksuid.New(),
			UserRef:    ref,
			RoleID:     role.ID,
			PracticeID: practice.ID,
			Email:      "signin@example.com",
			CreatedAt:  time.Now(),
		}))
		env.authMock.On("SignIn", mock.Anything, "signin@example.com", "s3cr3t-pw").
			Return(&auth.SignInResult{
				AuthenticatedUser: auth.AuthenticatedUser{ID: ref, Email: "signin@example.com"},
				Session: auth.Session{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					TokenType:    "bearer",
					ExpiresIn:    3600,
					ExpiresAt:    1234567890,
				},
			}, nil).Once()

		w := env.do(http.MethodPost, "/v1/users/signin", "", contracts.SignInRequest{
			Email:    "signin@example.com",
			Password: "s3cr3t-pw",
		})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.SignInResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "signin@example.com", resp.Email)
		require.Equal(t, states.UserStateActive, resp.EntityState)
		require.Equal(t, practice.ID, resp.PracticeID)
		require.Equal(t, practice.ID, resp.Practice.ID)
		require.Equal(t, practice.Name, resp.Practice.Name)
		require.Equal(t, "access-token", resp.AccessToken)
		require.Equal(t, "bearer", resp.TokenType)
		require.Equal(t, 3600, resp.ExpiresIn)
		require.Equal(t, int64(1234567890), resp.ExpiresAt)
		require.NotContains(t, w.Body.String(), "refresh-token")
		require.NotContains(t, w.Body.String(), "refresh_token")

		cookie := responseCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		require.Equal(t, "refresh-token", cookie.Value)
		require.True(t, cookie.HttpOnly)
		require.True(t, cookie.Secure)
		require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
		require.Equal(t, "/v1/users", cookie.Path)
		require.Empty(t, cookie.Domain)
		require.Greater(t, cookie.MaxAge, 0)
	})

	t.Run("400_when_the_email_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users/signin", "", contracts.SignInRequest{Password: "s3cr3t-pw"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_the_email_is_malformed", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users/signin", "", contracts.SignInRequest{
			Email:    "not-an-email",
			Password: "s3cr3t-pw",
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("401_when_the_credentials_are_invalid", func(t *testing.T) {
		env := newE2E(t)
		env.authMock.On("SignIn", mock.Anything, "signin@example.com", "wrong-pw").
			Return(nil, rerror.New(errors.New("invalid credentials")).WithKind(rerror.Permission)).Once()

		w := env.do(http.MethodPost, "/v1/users/signin", "", contracts.SignInRequest{
			Email:    "signin@example.com",
			Password: "wrong-pw",
		})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

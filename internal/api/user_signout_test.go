package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_SignOut(t *testing.T) {
	t.Run("204_revokes_the_session_and_clears_the_cookie", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		env.authMock.On("SignOut", mock.Anything, token).Return(nil).Once()

		w := env.do(http.MethodPost, "/v1/users/signout", token, nil)
		require.Equal(t, http.StatusNoContent, w.Code)

		cookie := responseCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		require.Empty(t, cookie.Value)
		require.Less(t, cookie.MaxAge, 0)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users/signout", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/stretchr/testify/require"
)

func Test_GetUser(t *testing.T) {
	t.Run("200_returns_the_authenticated_user_with_their_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		w := env.do(http.MethodGet, "/v1/users/me", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetUserResponse
		decodeBody(t, w, &resp)
		require.NotEmpty(t, resp.ID)
		require.NotEmpty(t, resp.Email)
		require.Equal(t, states.UserStateActive, resp.EntityState)
		require.Equal(t, practice.ID, resp.PracticeID)
		require.Equal(t, practice.ID, resp.Practice.ID)
		require.Equal(t, practice.Name, resp.Practice.Name)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/users/me", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

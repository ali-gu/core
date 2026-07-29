package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_GetPractice(t *testing.T) {
	t.Run("200_returns_the_authenticated_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		w := env.do(http.MethodGet, "/v1/practices/"+practice.ID.String(), token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetPracticeResponse
		decodeBody(t, w, &resp)
		require.Equal(t, practice.ID, resp.ID)
		require.Equal(t, practice.Name, resp.Name)
	})

	t.Run("403_when_requesting_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherPractice := fixtures.NewPractice(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/practices/"+otherPractice.ID.String(), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, fmt.Sprintf("/v1/practices/%s", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

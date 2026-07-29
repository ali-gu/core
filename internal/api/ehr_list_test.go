package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/stretchr/testify/require"
)

func Test_GetEHR(t *testing.T) {
	t.Run("200_returns_the_supported_ehrs", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodGet, "/v1/ehrs", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetEHRResponse
		decodeBody(t, w, &resp)
		require.Equal(t, constants.NexHealthEHRs, resp.Data)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/ehrs", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

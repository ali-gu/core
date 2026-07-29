package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Health(t *testing.T) {
	t.Run("200_and_no_auth_required", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/health", "", nil)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "success", w.Body.String())
	})
}

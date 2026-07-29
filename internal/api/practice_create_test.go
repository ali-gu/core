package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/stretchr/testify/require"
)

func Test_CreatePractice(t *testing.T) {
	t.Run("201_creates_a_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/admin/practices", token, contracts.CreatePracticeRequest{Name: "Bright Smiles"})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreatePracticeResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "Bright Smiles", resp.Name)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/admin/practices", "", contracts.CreatePracticeRequest{Name: "Bright Smiles"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("403_when_the_user_is_not_a_super_admin", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/admin/practices", token, contracts.CreatePracticeRequest{Name: "Bright Smiles"})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("400_when_the_name_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/admin/practices", token, contracts.CreatePracticeRequest{})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

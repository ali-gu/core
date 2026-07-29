package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/stretchr/testify/require"
)

func Test_CreateRole(t *testing.T) {
	t.Run("201_creates_a_role", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()
		env.clearRoleType(constants.RoleTypeReader)

		w := env.do(http.MethodPost, "/v1/admin/roles", token, contracts.CreateRoleRequest{Type: constants.RoleTypeReader})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.CreateRoleResponse
		decodeBody(t, w, &resp)
		require.Equal(t, constants.RoleTypeReader, resp.Type)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/admin/roles", "", contracts.CreateRoleRequest{Type: constants.RoleTypeReader})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("403_when_the_user_is_not_a_super_admin", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/admin/roles", token, contracts.CreateRoleRequest{Type: constants.RoleTypeReader})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("400_when_the_type_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/admin/roles", token, contracts.CreateRoleRequest{})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

package biz_test

import (
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/stretchr/testify/require"
)

func Test_Role_Create(t *testing.T) {
	t.Run("success_admin", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		clearRoleType(t, cfg, constants.RoleTypeAdmin)

		role, err := bz.Role.Create(cfg.Ctx, cfg.DB, contracts.CreateRoleRequest{
			Type: constants.RoleTypeAdmin,
		})
		require.NoError(t, err)
		require.NotEmpty(t, role.ID)
		require.Equal(t, constants.RoleTypeAdmin, role.Type)
		require.Equal(t, states.RoleStateActive, role.EntityState)
		require.False(t, role.CreatedAt.IsZero())
		require.Nil(t, role.DisabledAt)
	})

	t.Run("success_super_admin", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		clearRoleType(t, cfg, constants.RoleTypeSuperAdmin)

		role, err := bz.Role.Create(cfg.Ctx, cfg.DB, contracts.CreateRoleRequest{
			Type: constants.RoleTypeSuperAdmin,
		})
		require.NoError(t, err)
		require.Equal(t, constants.RoleTypeSuperAdmin, role.Type)
	})

	t.Run("success_reader", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		clearRoleType(t, cfg, constants.RoleTypeReader)

		role, err := bz.Role.Create(cfg.Ctx, cfg.DB, contracts.CreateRoleRequest{
			Type: constants.RoleTypeReader,
		})
		require.NoError(t, err)
		require.Equal(t, constants.RoleTypeReader, role.Type)
	})

	t.Run("error_when_type_already_exists", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Role.Create(cfg.Ctx, cfg.DB, contracts.CreateRoleRequest{Type: constants.RoleTypeAdmin})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("error_with_invalid_role_type", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Role.Create(cfg.Ctx, cfg.DB, contracts.CreateRoleRequest{
			Type: constants.RoleType("NOT_A_REAL_ROLE"),
		})
		require.Error(t, err)
	})

	t.Run("error_with_empty_role_type", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Role.Create(cfg.Ctx, cfg.DB, contracts.CreateRoleRequest{})
		require.Error(t, err)
	})
}

package biz_test

import (
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_Practice_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		practice, err := bz.Practice.Create(cfg.Ctx, cfg.DB, contracts.CreatePracticeRequest{
			Name: "foo_practice",
		})
		require.NoError(t, err)
		require.NotEmpty(t, practice.ID)
		require.Equal(t, "foo_practice", practice.Name)
		require.Equal(t, states.PracticeStateCreated, practice.EntityState)
		require.False(t, practice.CreatedAt.IsZero())
		require.Nil(t, practice.UpdatedAt)
		require.Nil(t, practice.DisabledAt)
	})

	t.Run("success_allows_duplicate_names", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		first, err := bz.Practice.Create(cfg.Ctx, cfg.DB, contracts.CreatePracticeRequest{Name: "dup_practice"})
		require.NoError(t, err)

		second, err := bz.Practice.Create(cfg.Ctx, cfg.DB, contracts.CreatePracticeRequest{Name: "dup_practice"})
		require.NoError(t, err)

		require.NotEqual(t, first.ID, second.ID)
	})

	t.Run("success_with_empty_name", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		practice, err := bz.Practice.Create(cfg.Ctx, cfg.DB, contracts.CreatePracticeRequest{Name: ""})
		require.NoError(t, err)
		require.Equal(t, "", practice.Name)
	})
}

func Test_Practice_GetByID(t *testing.T) {
	t.Run("success_returns_the_practice", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		created := fixtures.NewPractice(t, cfg, bz)

		practice, err := bz.Practice.GetByID(cfg.Ctx, cfg.DB, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, practice.ID)
		require.Equal(t, created.Name, practice.Name)
		require.Equal(t, states.PracticeStateCreated, practice.EntityState)
	})

	t.Run("error_when_the_practice_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Practice.GetByID(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})
}

func Test_Practice_Update(t *testing.T) {
	t.Run("success_updates_all_fields", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		updated, err := bz.Practice.Update(cfg.Ctx, cfg.DB, practice.ID, contracts.UpdatePracticeRequest{
			Email:   ptr.To("ops@example.com"),
			ZipCode: ptr.To("95014"),
			Website: ptr.To("www.example.com"),
		})
		require.NoError(t, err)
		require.NotNil(t, updated.Email)
		require.Equal(t, "ops@example.com", *updated.Email)
		require.NotNil(t, updated.ZipCode)
		require.Equal(t, "95014", *updated.ZipCode)
		require.NotNil(t, updated.Website)
		require.Equal(t, "www.example.com", *updated.Website)
		require.NotNil(t, updated.UpdatedAt)
	})

	t.Run("success_updates_only_the_given_field", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz,
			fixtures.WithPracticeEmail("ops@example.com"),
			fixtures.WithPracticeZipCode("95014"),
			fixtures.WithPracticeWebsite("www.example.com"),
		)

		updated, err := bz.Practice.Update(cfg.Ctx, cfg.DB, practice.ID, contracts.UpdatePracticeRequest{
			Website: ptr.To("www.new-example.com"),
		})
		require.NoError(t, err)
		require.NotNil(t, updated.Email)
		require.Equal(t, "ops@example.com", *updated.Email)
		require.NotNil(t, updated.ZipCode)
		require.Equal(t, "95014", *updated.ZipCode)
		require.NotNil(t, updated.Website)
		require.Equal(t, "www.new-example.com", *updated.Website)
	})

	t.Run("success_no_op_when_no_fields_are_set", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz, fixtures.WithPracticeEmail("ops@example.com"))

		updated, err := bz.Practice.Update(cfg.Ctx, cfg.DB, practice.ID, contracts.UpdatePracticeRequest{})
		require.NoError(t, err)
		require.NotNil(t, updated.Email)
		require.Equal(t, "ops@example.com", *updated.Email)
	})

	t.Run("error_when_the_practice_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Practice.Update(cfg.Ctx, cfg.DB, ksuid.New(), contracts.UpdatePracticeRequest{
			Email: ptr.To("ops@example.com"),
		})
		require.Error(t, err)
	})
}

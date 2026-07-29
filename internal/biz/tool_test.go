package biz_test

import (
	"errors"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/tool"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Tool_Sync(t *testing.T) {
	t.Run("success_creates_new_updates_existing_and_skips_unknown_kinds", func(t *testing.T) {
		toolMock := tool.NewMockITool(t)

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:    storage.NewStorage(),
			TelnyxTool: toolMock,
		})

		existing := fixtures.NewTool(t, cfg, bz,
			fixtures.WithToolKind(constants.ToolKindFindAppointment),
			fixtures.WithToolRef("stale_find_appointment_ref"),
			fixtures.WithToolConfig(map[string]any{"url": "https://example.com/stale"}),
		)

		toolMock.On("List", mock.Anything).Return([]tool.ListToolsResult{
			{ID: "telnyx_book_appointment", Type: "webhook", DisplayName: constants.ToolKindBookAppointment.String(), Config: map[string]any{"url": "https://example.com/book"}},
			{ID: "telnyx_hang_up", Type: "hangup", DisplayName: constants.ToolKindHangUp.String(), Config: map[string]any{}},
			{ID: "telnyx_find_appointment_updated", Type: "webhook", DisplayName: constants.ToolKindFindAppointment.String(), Config: map[string]any{"url": "https://example.com/find-updated"}},
			{ID: "telnyx_unrelated", Type: "function", DisplayName: "some_unrelated_tool", Config: map[string]any{}},
		}, nil).Once()

		tools, err := bz.Tool.Sync(cfg.Ctx, cfg.DB)
		require.NoError(t, err)

		byKind := make(map[constants.ToolKind]storage.Tool, len(tools))
		for _, tl := range tools {
			byKind[tl.Kind] = tl
		}

		require.Contains(t, byKind, constants.ToolKindBookAppointment)
		require.Equal(t, "telnyx_book_appointment", byKind[constants.ToolKindBookAppointment].ToolRef)

		require.Contains(t, byKind, constants.ToolKindHangUp)
		require.Equal(t, "telnyx_hang_up", byKind[constants.ToolKindHangUp].ToolRef)

		require.Contains(t, byKind, constants.ToolKindFindAppointment)
		updated := byKind[constants.ToolKindFindAppointment]
		require.Equal(t, existing.ID, updated.ID)
		require.Equal(t, "telnyx_find_appointment_updated", updated.ToolRef)
		require.Equal(t, "https://example.com/find-updated", updated.Config["url"])

		for _, tl := range tools {
			require.NotEqual(t, "some_unrelated_tool", tl.Kind.String())
		}
	})

	t.Run("error_when_telnyx_list_fails", func(t *testing.T) {
		toolMock := tool.NewMockITool(t)
		toolMock.On("List", mock.Anything).Return(nil, errors.New("telnyx unavailable")).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:    storage.NewStorage(),
			TelnyxTool: toolMock,
		})

		_, err := bz.Tool.Sync(cfg.Ctx, cfg.DB)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telnyx unavailable")
	})
}

package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/tool"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_CreateTool(t *testing.T) {
	t.Run("200_syncs_matching_tools_from_telnyx", func(t *testing.T) {
		deps, _ := testutils.SetupBiz(t)
		toolMock := tool.NewMockITool(t)
		toolMock.On("List", mock.Anything).Return([]tool.ListToolsResult{
			{ID: "telnyx_book_appointment", Type: "webhook", DisplayName: constants.ToolKindBookAppointment.String(), Config: map[string]any{"url": "https://example.com/book"}},
			{ID: "telnyx_unrelated", Type: "function", DisplayName: "some_unrelated_tool", Config: map[string]any{}},
		}, nil).Once()
		deps.TelnyxTool = toolMock

		env := newE2EWithDeps(t, deps)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/admin/tools", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetToolsResponse
		decodeBody(t, w, &resp)

		var found *contracts.Tool
		for i, tl := range resp.Data {
			if tl.Kind == constants.ToolKindBookAppointment {
				found = &resp.Data[i]
			}
			require.NotEqual(t, "some_unrelated_tool", string(tl.Kind))
		}
		require.NotNil(t, found, "expected a synced book_appointment tool")
		require.Equal(t, "https://example.com/book", found.Config["url"])
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/admin/tools", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("403_when_the_user_is_not_a_super_admin", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/admin/tools", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

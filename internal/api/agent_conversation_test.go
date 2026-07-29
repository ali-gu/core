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

func Test_GetAgentConversation(t *testing.T) {
	t.Run("200_when_the_agent_belongs_to_the_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		agentRecord := fixtures.NewAgent(t, env.cfg, env.bz, fixtures.WithAgentPracticeID(practice.ID))

		w := env.do(http.MethodGet, "/v1/agents/"+agentRecord.ID.String()+"/conversations/conversation_id", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.GetAgentConversationResponse
		decodeBody(t, w, &resp)
		require.Empty(t, resp.Data.Messages)
		require.Empty(t, resp.Data.Audio)
		require.Empty(t, resp.Data.Insights)
	})

	t.Run("403_when_the_agent_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherAgent := fixtures.NewAgent(t, env.cfg, env.bz)

		w := env.do(http.MethodGet, "/v1/agents/"+otherAgent.ID.String()+"/conversations/conversation_id", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, fmt.Sprintf("/v1/agents/%s/conversations/conversation_id", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

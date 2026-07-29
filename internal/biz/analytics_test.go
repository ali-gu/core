package biz_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Analytics_GetAgentsAnalytics(t *testing.T) {
	t.Run("success_empty_when_practice_has_no_agents", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		performances, err := bz.Analytics.GetAgentsAnalytics(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Empty(t, performances)
	})

	t.Run("success_aggregates_conversation_counts_bookings_and_durations_per_agent", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.MatchedBy(func(params agent.CreateAgentParams) bool {
			return params.Name == "top_agent"
		})).Return(&agent.CreateAgentResult{ID: "telnyx_top_agent", Name: "top_agent"}, nil).Once()
		agentMock.On("Create", mock.Anything, mock.MatchedBy(func(params agent.CreateAgentParams) bool {
			return params.Name == "quiet_agent"
		})).Return(&agent.CreateAgentResult{ID: "telnyx_quiet_agent", Name: "quiet_agent"}, nil).Once()

		firstCallAt := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
		secondCallAt := time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC)

		agentMock.On("GetAnalytics", mock.Anything, agent.GetAnalyticsParams{AgentRef: "telnyx_top_agent"}).Return([]agent.ConversationAnalytics{
			{
				ConversationID: "conversation_1",
				CreatedAt:      firstCallAt,
				LastMessageAt:  firstCallAt.Add(60 * time.Second),
				Messages: []agent.ConversationMessage{
					{Role: "user", Text: "hi"},
					{Role: "tool", ToolCalls: []agent.ConversationToolCall{{ID: "tc_1", Name: "book_appointment"}}},
				},
			},
			{
				ConversationID: "conversation_2",
				CreatedAt:      secondCallAt,
				LastMessageAt:  secondCallAt.Add(30 * time.Second),
				Messages: []agent.ConversationMessage{
					{Role: "user", Text: "hi again"},
				},
			},
		}, nil).Once()
		agentMock.On("GetAnalytics", mock.Anything, agent.GetAnalyticsParams{AgentRef: "telnyx_quiet_agent"}).Return([]agent.ConversationAnalytics{
			{
				ConversationID: "conversation_3",
				CreatedAt:      firstCallAt,
				LastMessageAt:  firstCallAt.Add(15 * time.Second),
				Messages: []agent.ConversationMessage{
					{Role: "user", Text: "hello"},
				},
			},
		}, nil).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		practice := fixtures.NewPractice(t, cfg, bz)
		topAgent := activatedAgentNamed(t, cfg, bz, practice.ID, "top_agent")
		activatedAgentNamed(t, cfg, bz, practice.ID, "quiet_agent")

		performances, err := bz.Analytics.GetAgentsAnalytics(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Len(t, performances, 2)

		best := performances[0]
		require.Equal(t, topAgent.ID, best.Agent.ID)
		require.Equal(t, 2, best.ConversationCount)
		require.Equal(t, 1, best.BookingsMade)
		require.InDelta(t, 0.5, best.ConversionRate(), 0.0001)
		require.Equal(t, 60*time.Second, best.LongestConversation)
		require.Equal(t, 30*time.Second, best.ShortestConversation)
		require.NotNil(t, best.LastConversationAt)
		require.True(t, best.LastConversationAt.Equal(secondCallAt))

		quiet := performances[1]
		require.Equal(t, 1, quiet.ConversationCount)
		require.Equal(t, 0, quiet.BookingsMade)
		require.Zero(t, quiet.ConversionRate())
		require.Equal(t, 15*time.Second, quiet.LongestConversation)
		require.Equal(t, 15*time.Second, quiet.ShortestConversation)
	})

	t.Run("success_only_includes_agents_from_the_given_practice", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)
		fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentPracticeID(practice.ID))
		fixtures.NewAgent(t, cfg, bz)

		performances, err := bz.Analytics.GetAgentsAnalytics(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Len(t, performances, 1)
	})

	t.Run("error_when_telnyx_analytics_fails_for_any_agent", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.Anything).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()
		agentMock.On("GetAnalytics", mock.Anything, mock.Anything).Return(nil, errors.New("telnyx unavailable")).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		practice := fixtures.NewPractice(t, cfg, bz)
		activatedAgentNamed(t, cfg, bz, practice.ID, "foo_agent")

		_, err := bz.Analytics.GetAgentsAnalytics(cfg.Ctx, cfg.DB, practice.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telnyx unavailable")
	})
}

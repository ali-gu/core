package biz_test

import (
	"errors"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Conversation_LogConversation(t *testing.T) {
	t.Run("success_derives_outcome_from_tools_called", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)

		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("agent_ref_123"),
		)

		phoneNumber := fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		err := bz.Conversation.LogConversation(cfg.Ctx, cfg.DB, contracts.LogConversationRequest{
			AssistantID:     "agent_ref_123",
			ConversationRef: "conv_123",
			CallDuration:    90000,
			ToolsCalled: []string{
				constants.ToolKindFindAppointment.String(),
				constants.ToolKindBookAppointment.String(),
			},
		})
		require.NoError(t, err)

		list, err := cfg.Deps.Storage.Conversation.Get(cfg.Ctx, cfg.DB, storage.ConversationFilters{
			AgentID: &agentRecord.ID,
		})
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, agentRecord.ID, list[0].AgentID)
		require.Equal(t, phoneNumber.ID, list[0].PhoneNumberID)
		require.Equal(t, location.ID, list[0].LocationID)
		require.Equal(t, location.PracticeID, list[0].PracticeID)
		require.Equal(t, "conv_123", list[0].ConversationRef)
		require.Equal(t, int64(90000), list[0].Duration)
		require.Equal(t, "Appointment booked", list[0].Outcome)
	})

	t.Run("outcome_defaults_to_info_when_no_tools_called", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("agent_ref_123"),
		)
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		err := bz.Conversation.LogConversation(cfg.Ctx, cfg.DB, contracts.LogConversationRequest{
			AssistantID:     "agent_ref_123",
			ConversationRef: "conv_456",
		})
		require.NoError(t, err)

		list, err := cfg.Deps.Storage.Conversation.Get(cfg.Ctx, cfg.DB, storage.ConversationFilters{
			AgentID: &agentRecord.ID,
		})
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, "Info", list[0].Outcome)
	})

	t.Run("error_when_agent_ref_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		err := bz.Conversation.LogConversation(cfg.Ctx, cfg.DB, contracts.LogConversationRequest{
			AssistantID:     "missing_ref",
			ConversationRef: "conv_123",
		})
		require.Error(t, err)
	})

	t.Run("error_when_agent_has_no_location_assigned", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		unassignedAgent := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentRef("agent_ref_123"))
		require.Nil(t, unassignedAgent.LocationID)

		err := bz.Conversation.LogConversation(cfg.Ctx, cfg.DB, contracts.LogConversationRequest{
			AssistantID:     "agent_ref_123",
			ConversationRef: "conv_123",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "no location assigned")
	})
}

func Test_Conversation_GetConversation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.Anything).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()
		agentMock.On("GetConversation", mock.Anything, agent.GetConversationParams{
			AgentRef:       "telnyx_agent_id",
			ConversationID: "conversation_id",
		}).Return(&agent.ConversationAnalytics{
			ConversationID: "conversation_id",
			Status:         "completed",
			Metadata:       map[string]string{"call_control_id": "call_control_id"},
			Insights:       map[string]string{"insight_id": "sentiment: positive"},
			Messages: []agent.ConversationMessage{
				{Role: "user", Text: "hello"},
			},
		}, nil).Once()

		agentMock.On("GetRecordings", mock.Anything, agent.GetRecordingsParams{CallControlID: "call_control_id"}).Return([]agent.ConversationRecording{
			{ID: "recording_id", MP3URL: "https://example.com/recording.mp3"},
		}, nil).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		agentRecord := activatedAgent(t, cfg, bz)

		conversation, err := bz.Conversation.GetConversation(cfg.Ctx, cfg.DB, agentRecord.ID, "conversation_id")
		require.NoError(t, err)
		require.Equal(t, "conversation_id", conversation.ConversationID)
		require.Equal(t, "completed", conversation.Status)
		require.Len(t, conversation.Messages, 1)
		require.Len(t, conversation.Recordings, 1)
		require.Equal(t, "https://example.com/recording.mp3", conversation.Recordings[0].MP3URL)
	})

	t.Run("error_when_agent_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Conversation.GetConversation(cfg.Ctx, cfg.DB, ksuid.New(), "conversation_id")
		require.Error(t, err)
	})

	t.Run("forbidden_when_conversation_not_found_or_not_owned_by_agent", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.Anything).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()
		agentMock.On("GetConversation", mock.Anything, agent.GetConversationParams{
			AgentRef:       "telnyx_agent_id",
			ConversationID: "conversation_id",
		}).Return(nil, agent.ErrConversationNotFound).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		agentRecord := activatedAgent(t, cfg, bz)

		_, err := bz.Conversation.GetConversation(cfg.Ctx, cfg.DB, agentRecord.ID, "conversation_id")
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})
}

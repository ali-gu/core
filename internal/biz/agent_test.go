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
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Agent_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		agentRecord, err := bz.Agent.Create(cfg.Ctx, cfg.DB, practice.ID, contracts.CreateAgentRequest{
			Name: "foo_agent",
		})
		require.NoError(t, err)
		require.NotEmpty(t, agentRecord.ID)
		require.Equal(t, practice.ID, agentRecord.PracticeID)
		require.Empty(t, agentRecord.AgentRef)
		require.Equal(t, "foo_agent", agentRecord.Name)
		require.Equal(t, states.AgentStateCreated, agentRecord.EntityState)
		require.Nil(t, agentRecord.LocationID)
		require.Nil(t, agentRecord.PhoneNumberID)
	})

	t.Run("success_with_a_location", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))

		agentRecord, err := bz.Agent.Create(cfg.Ctx, cfg.DB, location.PracticeID, contracts.CreateAgentRequest{
			Name:       "foo_agent",
			LocationID: &location.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, agentRecord.LocationID)
		require.Equal(t, location.ID, *agentRecord.LocationID)
	})

	t.Run("success_with_a_phone_number", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)
		phoneNumber := fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		agentRecord, err := bz.Agent.Create(cfg.Ctx, cfg.DB, location.PracticeID, contracts.CreateAgentRequest{
			Name:          "foo_agent",
			PhoneNumberID: &phoneNumber.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, agentRecord.PhoneNumberID)
		require.Equal(t, phoneNumber.ID, *agentRecord.PhoneNumberID)
	})

	t.Run("error_when_location_is_not_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		location := fixtures.NewPendingLocation(t, cfg, bz)

		_, err := bz.Agent.Create(cfg.Ctx, cfg.DB, location.PracticeID, contracts.CreateAgentRequest{
			Name:       "foo_agent",
			LocationID: &location.ID,
		})
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})

	t.Run("error_when_phone_number_is_not_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practice.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550199",
		})
		require.NoError(t, err)

		_, err = bz.Agent.Create(cfg.Ctx, cfg.DB, practice.ID, contracts.CreateAgentRequest{
			Name:          "foo_agent",
			PhoneNumberID: &reserved.ID,
		})
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})

	t.Run("error_when_practice_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Agent.Create(cfg.Ctx, cfg.DB, ksuid.New(), contracts.CreateAgentRequest{
			Name: "foo_agent",
		})
		require.Error(t, err)
	})
}

func Test_Agent_Get(t *testing.T) {
	t.Run("success_empty", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		agents, err := bz.Agent.Get(cfg.Ctx, cfg.DB, ksuid.New())
		require.NoError(t, err)
		require.Empty(t, agents)
	})

	t.Run("success_returns_only_the_given_practices_agents", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)
		fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentPracticeID(practice.ID))
		fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentPracticeID(practice.ID))

		fixtures.NewAgent(t, cfg, bz)

		agents, err := bz.Agent.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Len(t, agents, 2)
		for _, a := range agents {
			require.Equal(t, practice.ID, a.PracticeID)
		}
	})

	t.Run("success_hydrates_assigned_location_and_phone_number", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
		phoneNumber := fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		agents, err := bz.Agent.Get(cfg.Ctx, cfg.DB, location.PracticeID)
		require.NoError(t, err)
		require.Len(t, agents, 1)
		require.NotNil(t, agents[0].Location)
		require.Equal(t, location.ID, agents[0].Location.ID)
		require.NotNil(t, agents[0].PhoneNumber)
		require.Equal(t, phoneNumber.ID, agents[0].PhoneNumber.ID)
	})
}

func Test_Agent_Update(t *testing.T) {
	t.Run("success_assigns_a_location", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)
		location := fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(agentRecord.PracticeID))
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			LocationID: &location.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.LocationID)
		require.Equal(t, location.ID, *updated.LocationID)
		require.NotNil(t, updated.UpdatedAt)
	})

	t.Run("success_reassigning_a_location_unassigns_the_previous_agent", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		firstAgent := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))

		secondAgent, err := bz.Agent.Create(cfg.Ctx, cfg.DB, location.PracticeID, contracts.CreateAgentRequest{
			Name: "second_agent",
		})
		require.NoError(t, err)

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, secondAgent.ID, contracts.UpdateAgentRequest{
			LocationID: &location.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.LocationID)
		require.Equal(t, location.ID, *updated.LocationID)

		previous, err := cfg.Deps.Storage.Agent.GetByID(cfg.Ctx, cfg.DB, firstAgent.ID)
		require.NoError(t, err)
		require.Nil(t, previous.LocationID)
	})

	t.Run("success_assigns_a_phone_number", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)
		phoneNumber := fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))
		agentRecord, err := bz.Agent.Create(cfg.Ctx, cfg.DB, location.PracticeID, contracts.CreateAgentRequest{
			Name: "foo_agent",
		})
		require.NoError(t, err)

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			PhoneNumberID: &phoneNumber.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.PhoneNumberID)
		require.Equal(t, phoneNumber.ID, *updated.PhoneNumberID)
	})

	t.Run("success_reassigning_a_phone_number_unassigns_the_previous_agent", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))

		phoneNumber := fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		secondAgent, err := bz.Agent.Create(cfg.Ctx, cfg.DB, location.PracticeID, contracts.CreateAgentRequest{
			Name: "second_agent",
		})
		require.NoError(t, err)

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, secondAgent.ID, contracts.UpdateAgentRequest{
			PhoneNumberID: &phoneNumber.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.PhoneNumberID)
		require.Equal(t, phoneNumber.ID, *updated.PhoneNumberID)

		previous, err := cfg.Deps.Storage.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Nil(t, previous.PhoneNumberID)
	})

	t.Run("success_assigns_location_and_phone_number_together", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		phoneNumber := fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentPracticeID(location.PracticeID))

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			LocationID:    &location.ID,
			PhoneNumberID: &phoneNumber.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.LocationID)
		require.Equal(t, location.ID, *updated.LocationID)
		require.NotNil(t, updated.PhoneNumberID)
		require.Equal(t, phoneNumber.ID, *updated.PhoneNumberID)
	})

	t.Run("success_updates_the_name", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			Name: ptr.To("new_name"),
		})
		require.NoError(t, err)
		require.Equal(t, "new_name", updated.Name)
		require.NotNil(t, updated.UpdatedAt)
	})

	t.Run("success_no_op_when_no_fields_are_set", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		updated, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{})
		require.NoError(t, err)
		require.Equal(t, agentRecord.ID, updated.ID)
		require.Nil(t, updated.UpdatedAt)
	})

	t.Run("error_when_location_is_not_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		location := fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(agentRecord.PracticeID))

		_, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			LocationID: &location.ID,
		})
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())

		unchanged, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Nil(t, unchanged.LocationID)
	})

	t.Run("error_when_phone_number_is_not_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, agentRecord.PracticeID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550199",
		})
		require.NoError(t, err)

		_, err = bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			PhoneNumberID: &reserved.ID,
		})
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())

		unchanged, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Nil(t, unchanged.PhoneNumberID)
	})

	t.Run("error_when_agent_is_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)
		setAgentEntityState(t, cfg, bz, agentRecord.ID, states.AgentStateActive)

		_, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			Name: ptr.To("new_name"),
		})
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())

		unchanged, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, agentRecord.Name, unchanged.Name)
	})

	t.Run("error_when_agent_is_disabled", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)
		setAgentEntityState(t, cfg, bz, agentRecord.ID, states.AgentStateDisabled)

		_, err := bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			Name: ptr.To("new_name"),
		})
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())

		unchanged, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, agentRecord.Name, unchanged.Name)
	})

	t.Run("error_when_agent_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Agent.Update(cfg.Ctx, cfg.DB, ksuid.New(), contracts.UpdateAgentRequest{})
		require.Error(t, err)
	})
}

func Test_Agent_Activate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))

		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		updated, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateActive, updated.EntityState)
		require.NotNil(t, updated.UpdatedAt)
		require.Equal(t, "telnyx_agent_id", ptr.From(updated.AgentRef))
		require.Equal(t, "telnyx_agent_name", updated.Name)
	})

	t.Run("success_reactivates_without_calling_telnyx_when_agent_ref_already_set", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.Anything).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)

		disabled, err := bz.Agent.Disable(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateDisabled, disabled.EntityState)
		require.NotEmpty(t, disabled.AgentRef)

		reactivated, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateActive, reactivated.EntityState)
		require.Equal(t, disabled.AgentRef, reactivated.AgentRef)
	})

	t.Run("success_assigns_the_phone_number_to_the_telnyx_agent_on_create", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.MatchedBy(func(params agent.CreateAgentParams) bool {
			return params.PhoneNumberIDRef == "telnyx_number_ref"
		})).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
		fixtures.NewPhoneNumber(t, cfg, bz,
			fixtures.WithPhoneNumberLocationID(location.ID), fixtures.WithPhoneNumberIDRef("telnyx_number_ref"))

		updated, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateActive, updated.EntityState)
	})

	t.Run("success_active_tools_tool_refs_are_attached_to_the_telnyx_create_call", func(t *testing.T) {
		var capturedToolIDs []string
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.MatchedBy(func(params agent.CreateAgentParams) bool {
			capturedToolIDs = params.ToolIDs
			return true
		})).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		fixtures.NewTool(t, cfg, bz, fixtures.WithToolRef("active_tool_ref"))

		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)

		require.Contains(t, capturedToolIDs, "active_tool_ref")
	})

	t.Run("success_disabled_tools_are_excluded_from_the_telnyx_create_call", func(t *testing.T) {
		var capturedToolIDs []string
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.MatchedBy(func(params agent.CreateAgentParams) bool {
			capturedToolIDs = params.ToolIDs
			return true
		})).Return(&agent.CreateAgentResult{
			ID:   "telnyx_agent_id",
			Name: "telnyx_agent_name",
		}, nil).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		fixtures.NewTool(t, cfg, bz, fixtures.WithToolRef("disabled_tool_ref"), fixtures.WithToolDisabled())

		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.NotContains(t, capturedToolIDs, "disabled_tool_ref")
	})

	t.Run("error_when_telnyx_create_fails", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("telnyx unavailable")).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telnyx unavailable")

		unchanged, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateCreated, unchanged.EntityState)
		require.Empty(t, unchanged.AgentRef)
	})

	t.Run("error_when_no_location_assigned", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})

	t.Run("error_when_no_phone_number_assigned", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)
	})

	t.Run("error_when_assigned_location_is_not_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		location := fixtures.NewPendingLocation(t, cfg, bz)
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentPracticeID(location.PracticeID))

		agentRecord.AssignLocation(location.ID)
		require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, agentRecord))

		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)
	})

	t.Run("error_when_assigned_phone_number_is_not_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, agentRecord.PracticeID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550199",
		})
		require.NoError(t, err)

		agentRecord.AssignPhoneNumber(reserved.ID)
		require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, agentRecord))

		_, err = bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)
	})

	t.Run("error_when_assigned_phone_number_has_no_telnyx_ref", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))

		fixtures.NewPhoneNumber(t, cfg, bz,
			fixtures.WithPhoneNumberLocationID(location.ID), fixtures.WithPhoneNumberNoIDRef())

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})

	t.Run("error_when_agent_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})
}

func Test_Agent_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		require.NoError(t, bz.Agent.Delete(cfg.Ctx, cfg.DB, agentRecord.ID))

		agents, err := bz.Agent.Get(cfg.Ctx, cfg.DB, agentRecord.PracticeID)
		require.NoError(t, err)
		require.Empty(t, agents)
	})

	t.Run("error_when_agent_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		err := bz.Agent.Delete(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})

	t.Run("success_keeps_the_shared_telnyx_assistant_for_a_test_agent", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
			Landscape:                constants.LandscapeTest,
		})
		practice := fixtures.NewPractice(t, cfg, bz)
		testAgent, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)

		require.NoError(t, bz.Agent.Delete(cfg.Ctx, cfg.DB, testAgent.ID))
		agentMock.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)

		agents, err := bz.Agent.Get(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.Empty(t, agents)
	})

	t.Run("error_when_telnyx_delete_fails_leaves_the_agent_in_place", func(t *testing.T) {
		agentMock := agent.NewMockIAgent(t)
		agentMock.On("Delete", mock.Anything, mock.Anything).Return(errors.New("telnyx unavailable")).Once()

		cfg, bz := testutils.BasicSetupWithDeps(t, biz.Dependencies{
			Storage:                  storage.NewStorage(),
			TelnyxAgent:              agentMock,
			TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
			EHR:                      ehr.NewMockIEHR(t),
		})
		agentRecord := fixtures.NewAgent(t, cfg, bz)

		err := bz.Agent.Delete(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telnyx unavailable")

		agents, err := bz.Agent.Get(cfg.Ctx, cfg.DB, agentRecord.PracticeID)
		require.NoError(t, err)
		require.Len(t, agents, 1)
	})
}

func activatedAgent(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz) storage.Agent {
	t.Helper()

	location := fixtures.NewPendingLocation(t, cfg, bz)
	fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
	agentRecord := fixtures.NewAgent(t, cfg, bz, fixtures.WithAgentLocationID(location.ID))
	fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

	activated, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
	require.NoError(t, err)
	return *activated
}

func activatedAgentNamed(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, practiceID ksuid.KSUID, name string) storage.Agent {
	t.Helper()

	location := fixtures.NewPendingLocation(t, cfg, bz, fixtures.WithLocationPracticeID(practiceID))
	fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
	agentRecord := fixtures.NewAgent(t, cfg, bz,
		fixtures.WithAgentPracticeID(practiceID), fixtures.WithAgentName(name), fixtures.WithAgentLocationID(location.ID))
	fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

	activated, err := bz.Agent.Activate(cfg.Ctx, cfg.DB, agentRecord.ID)
	require.NoError(t, err)
	return *activated
}

func setAgentEntityState(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, id ksuid.KSUID, state states.AgentState) {
	t.Helper()

	agentRecord, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, id)
	require.NoError(t, err)

	agentRecord.EntityState = state
	require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, *agentRecord))
}

func Test_Agent_Disable(t *testing.T) {
	t.Run("success_disabling_an_active_agent", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)
		setAgentEntityState(t, cfg, bz, agentRecord.ID, states.AgentStateActive)

		disabled, err := bz.Agent.Disable(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateDisabled, disabled.EntityState)
		require.NotNil(t, disabled.UpdatedAt)
	})

	t.Run("error_disabling_an_agent_that_isnt_active_yet", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)
		setAgentEntityState(t, cfg, bz, agentRecord.ID, states.AgentStateCreated)

		_, err := bz.Agent.Disable(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())

		agentRecord2, err := bz.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
		require.Equal(t, states.AgentStateCreated, agentRecord2.EntityState)
	})

	t.Run("error_disabling_an_already_disabled_agent", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		agentRecord := fixtures.NewAgent(t, cfg, bz)
		setAgentEntityState(t, cfg, bz, agentRecord.ID, states.AgentStateActive)

		_, err := bz.Agent.Disable(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)

		_, err = bz.Agent.Disable(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.Error(t, err)

		var rerr *rerror.Error
		require.True(t, errors.As(err, &rerr))
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})

	t.Run("error_when_agent_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Agent.Disable(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})
}

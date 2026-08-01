package biz_test

import (
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_TestAgent_Create(t *testing.T) {
	testAgent := constants.TestAgentFor(constants.LandscapeTest)

	t.Run("success_connects_the_agent_to_the_landscape_telnyx_assistant", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		agentRecord, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)

		require.Equal(t, practice.ID, agentRecord.PracticeID)
		require.Equal(t, constants.TestAgentName, agentRecord.Name)
		require.Equal(t, states.AgentStateActive, agentRecord.EntityState)
		require.Equal(t, testAgent.AgentRef, ptr.From(agentRecord.AgentRef))
		require.NotNil(t, agentRecord.LocationID)
		require.NotNil(t, agentRecord.PhoneNumberID)
	})

	t.Run("success_creates_an_active_location_with_the_nexhealth_ehr", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		agentRecord, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)

		location, err := cfg.Deps.Storage.Location.GetByID(cfg.Ctx, cfg.DB, *agentRecord.LocationID)
		require.NoError(t, err)
		require.Equal(t, states.LocationStateActive, location.EntityState)
		require.Equal(t, constants.TestAgentLocationAddress, location.Address)
		require.Equal(t, practice.ID, location.PracticeID)

		ehrRecord, err := cfg.Deps.Storage.EHR.GetByLocationID(cfg.Ctx, cfg.DB, location.ID)
		require.NoError(t, err)
		require.Equal(t, constants.EHRNexHealth, ehrRecord.Type)
		require.Equal(t, testAgent.EHRSubdomain, ehrRecord.Subdomain)
		require.Equal(t, testAgent.EHRLocationRef, ptr.From(ehrRecord.LocationRef))
	})

	t.Run("success_creates_an_active_phone_number_with_its_telnyx_ref", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		agentRecord, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)

		phoneNumber, err := cfg.Deps.Storage.PhoneNumber.GetByID(cfg.Ctx, cfg.DB, *agentRecord.PhoneNumberID)
		require.NoError(t, err)
		require.Equal(t, states.PhoneNumberStateActive, phoneNumber.EntityState)
		require.Equal(t, testAgent.PhoneNumber, phoneNumber.PhoneNumber)
		require.Equal(t, testAgent.PhoneNumberRef, ptr.From(phoneNumber.PhoneNumberIDRef))
		require.Equal(t, practice.ID, phoneNumber.PracticeID)
	})

	t.Run("error_when_the_practice_already_has_a_test_agent", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		_, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)

		_, err = bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.Error(t, err)

		var rerr *rerror.Error
		require.ErrorAs(t, err, &rerr)
		require.Equal(t, rerror.Forbidden, rerr.Kind())
	})

	t.Run("success_shares_the_telnyx_resources_with_another_practices_test_agent", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)
		other := fixtures.NewPractice(t, cfg, bz)

		first, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)

		second, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, other.ID)
		require.NoError(t, err)

		require.NotEqual(t, first.ID, second.ID)
		require.Equal(t, other.ID, second.PracticeID)
		require.Equal(t, ptr.From(first.AgentRef), ptr.From(second.AgentRef))

		require.NotEqual(t, *first.LocationID, *second.LocationID)
		require.NotEqual(t, *first.PhoneNumberID, *second.PhoneNumberID)

		firstPhoneNumber, err := cfg.Deps.Storage.PhoneNumber.GetByID(cfg.Ctx, cfg.DB, *first.PhoneNumberID)
		require.NoError(t, err)
		secondPhoneNumber, err := cfg.Deps.Storage.PhoneNumber.GetByID(cfg.Ctx, cfg.DB, *second.PhoneNumberID)
		require.NoError(t, err)
		require.Equal(t, firstPhoneNumber.PhoneNumber, secondPhoneNumber.PhoneNumber)
		require.Equal(t, ptr.From(firstPhoneNumber.PhoneNumberIDRef), ptr.From(secondPhoneNumber.PhoneNumberIDRef))
	})

	t.Run("error_when_the_practice_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, ksuid.New())
		require.Error(t, err)
	})

	t.Run("success_reuses_and_reactivates_the_existing_test_resources", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		first, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.NoError(t, cfg.Deps.Storage.Agent.Delete(cfg.Ctx, cfg.DB, first.ID))

		location, err := cfg.Deps.Storage.Location.GetByID(cfg.Ctx, cfg.DB, *first.LocationID)
		require.NoError(t, err)
		location.LocationToDisabled()
		require.NoError(t, cfg.Deps.Storage.Location.Update(cfg.Ctx, cfg.DB, *location))

		phoneNumber, err := cfg.Deps.Storage.PhoneNumber.GetByID(cfg.Ctx, cfg.DB, *first.PhoneNumberID)
		require.NoError(t, err)
		phoneNumber.PhoneNumberToDisabled()
		require.NoError(t, cfg.Deps.Storage.PhoneNumber.Update(cfg.Ctx, cfg.DB, *phoneNumber))

		second, err := bz.TestAgent.Create(cfg.Ctx, cfg.DB, practice.ID)
		require.NoError(t, err)
		require.NotEqual(t, first.ID, second.ID)
		require.Equal(t, *first.LocationID, *second.LocationID)
		require.Equal(t, *first.PhoneNumberID, *second.PhoneNumberID)

		locations, err := cfg.Deps.Storage.Location.Get(cfg.Ctx, cfg.DB, storage.LocationFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, locations, 1)
		require.Equal(t, states.LocationStateActive, locations[0].EntityState)

		phoneNumbers, err := cfg.Deps.Storage.PhoneNumber.Get(cfg.Ctx, cfg.DB, storage.PhoneNumberFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, phoneNumbers, 1)
		require.Equal(t, states.PhoneNumberStateActive, phoneNumbers[0].EntityState)
	})
}

func Test_TestAgentFor(t *testing.T) {
	t.Run("returns_the_prod_resources_for_prod", func(t *testing.T) {
		prod := constants.TestAgentFor(constants.LandscapeProd)
		require.Equal(t, "prod-dentist", prod.EHRSubdomain)
		require.Equal(t, "1234", prod.EHRLocationRef)
	})

	t.Run("returns_the_same_resources_for_every_other_landscape", func(t *testing.T) {
		local := constants.TestAgentFor(constants.LandscapeLocal)
		require.Equal(t, local, constants.TestAgentFor(constants.LandscapeDev))
		require.Equal(t, local, constants.TestAgentFor(constants.LandscapeTest))
		require.NotEqual(t, local, constants.TestAgentFor(constants.LandscapeProd))
	})
}

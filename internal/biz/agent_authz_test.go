package biz_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func activeLocationForPractice(t *testing.T, cfg *testutils.TestConfig, practiceID ksuid.KSUID) storage.Location {
	t.Helper()

	location := storage.Location{
		EntityBase: storage.EntityBase[states.LocationState]{EntityState: states.LocationStateActive},
		ID:         ksuid.New(),
		Address:    "1 Cross St",
		PracticeID: practiceID,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, cfg.Deps.Storage.Location.Create(cfg.Ctx, cfg.DB, location))
	return location
}

func activePhoneNumberForPractice(t *testing.T, cfg *testutils.TestConfig, practiceID ksuid.KSUID) storage.PhoneNumber {
	t.Helper()

	phoneNumber := storage.PhoneNumber{
		EntityBase:       storage.EntityBase[states.PhoneNumberState]{EntityState: states.PhoneNumberStateActive},
		ID:               ksuid.New(),
		PhoneNumber:      "+15555550" + ksuid.New().String()[:3],
		PhoneNumberIDRef: ptr.To("telnyx_ref_" + ksuid.New().String()),
		PracticeID:       practiceID,
		CreatedAt:        time.Now(),
	}
	require.NoError(t, cfg.Deps.Storage.PhoneNumber.Create(cfg.Ctx, cfg.DB, phoneNumber))
	return phoneNumber
}

func requireForbidden(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	var rerr *rerror.Error
	require.True(t, errors.As(err, &rerr))
	require.Equal(t, rerror.Forbidden, rerr.Kind())
}

func Test_Agent_Create_RejectsCrossPracticeLocation(t *testing.T) {
	cfg, bz := testutils.BasicSetup(t)

	victim := fixtures.NewPractice(t, cfg, bz)
	victimLocation := activeLocationForPractice(t, cfg, victim.ID)

	attacker := fixtures.NewPractice(t, cfg, bz)

	_, err := bz.Agent.Create(cfg.Ctx, cfg.DB, attacker.ID, contracts.CreateAgentRequest{
		Name:       "attacker_agent",
		LocationID: &victimLocation.ID,
	})
	requireForbidden(t, err)
}

func Test_Agent_Create_RejectsCrossPracticePhoneNumber(t *testing.T) {
	cfg, bz := testutils.BasicSetup(t)

	victim := fixtures.NewPractice(t, cfg, bz)
	victimPhone := activePhoneNumberForPractice(t, cfg, victim.ID)

	attacker := fixtures.NewPractice(t, cfg, bz)

	_, err := bz.Agent.Create(cfg.Ctx, cfg.DB, attacker.ID, contracts.CreateAgentRequest{
		Name:          "attacker_agent",
		PhoneNumberID: &victimPhone.ID,
	})
	requireForbidden(t, err)
}

func Test_Agent_Update_RejectsCrossPracticeLocation(t *testing.T) {
	cfg, bz := testutils.BasicSetup(t)

	victim := fixtures.NewPractice(t, cfg, bz)
	victimLocation := activeLocationForPractice(t, cfg, victim.ID)

	attacker := fixtures.NewPractice(t, cfg, bz)
	attackerAgent, err := bz.Agent.Create(cfg.Ctx, cfg.DB, attacker.ID, contracts.CreateAgentRequest{
		Name: "attacker_agent",
	})
	require.NoError(t, err)

	_, err = bz.Agent.Update(cfg.Ctx, cfg.DB, attackerAgent.ID, contracts.UpdateAgentRequest{
		LocationID: &victimLocation.ID,
	})
	requireForbidden(t, err)
}

func Test_Agent_Update_RejectsCrossPracticePhoneNumber(t *testing.T) {
	cfg, bz := testutils.BasicSetup(t)

	victim := fixtures.NewPractice(t, cfg, bz)
	victimPhone := activePhoneNumberForPractice(t, cfg, victim.ID)

	attacker := fixtures.NewPractice(t, cfg, bz)
	attackerAgent, err := bz.Agent.Create(cfg.Ctx, cfg.DB, attacker.ID, contracts.CreateAgentRequest{
		Name: "attacker_agent",
	})
	require.NoError(t, err)

	_, err = bz.Agent.Update(cfg.Ctx, cfg.DB, attackerAgent.ID, contracts.UpdateAgentRequest{
		PhoneNumberID: &victimPhone.ID,
	})
	requireForbidden(t, err)
}

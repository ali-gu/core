package biz

import (
	"context"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type TestAgent struct {
	*Biz

	storage   storage.Storage
	landscape constants.Landscape
}

type ITestAgent interface {
	Create(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) (*storage.Agent, error)
}

var _ ITestAgent = (*TestAgent)(nil)

func (t *TestAgent) Create(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) (*storage.Agent, error) {
	testAgent := constants.TestAgentFor(t.landscape)

	existing, err := t.storage.Agent.Get(ctx, db, storage.AgentFilters{
		PracticeID: &practiceID,
		AgentRef:   &testAgent.AgentRef,
	})
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	if len(existing) > 0 {
		return nil, rerror.NewMessage("this practice already has a test agent", rerror.Forbidden)
	}

	if _, err = t.storage.Practice.GetByID(ctx, db, practiceID); err != nil {
		return nil, rerror.Wrap(err)
	}

	locationID, err := t.location(ctx, db, practiceID, testAgent)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	phoneNumberID, err := t.phoneNumber(ctx, db, practiceID, testAgent)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	agentID := ksuid.New()
	if err = t.storage.Agent.Create(ctx, db, storage.Agent{
		EntityBase:    storage.EntityBase[states.AgentState]{EntityState: states.AgentStateActive},
		ID:            agentID,
		PracticeID:    practiceID,
		AgentRef:      ptr.To(testAgent.AgentRef),
		Name:          constants.TestAgentName,
		LocationID:    &locationID,
		PhoneNumberID: &phoneNumberID,
		CreatedAt:     time.Now(),
	}); err != nil {
		return nil, rerror.Wrap(err)
	}

	return t.storage.Agent.GetByID(ctx, db, agentID)
}

func (t *TestAgent) location(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, testAgent constants.TestAgent) (ksuid.KSUID, error) {
	var noop ksuid.KSUID

	locations, err := t.storage.Location.Get(ctx, db, storage.LocationFilters{PracticeID: &practiceID})
	if err != nil {
		return noop, rerror.Wrap(err)
	}

	for _, location := range locations {
		if location.EHR == nil ||
			location.EHR.Subdomain != testAgent.EHRSubdomain ||
			ptr.From(location.EHR.LocationRef) != testAgent.EHRLocationRef {
			continue
		}

		if location.EntityState != states.LocationStateActive {
			location.LocationToActive()
			if err = t.storage.Location.Update(ctx, db, location); err != nil {
				return noop, rerror.Wrap(err)
			}
		}
		return location.ID, nil
	}

	locationID := ksuid.New()
	if err = t.storage.Location.Create(ctx, db, storage.Location{
		EntityBase: storage.EntityBase[states.LocationState]{EntityState: states.LocationStateActive},
		ID:         locationID,
		Address:    constants.TestAgentLocationAddress,
		PracticeID: practiceID,
		CreatedAt:  time.Now(),
	}); err != nil {
		return noop, rerror.Wrap(err)
	}

	if err = t.storage.EHR.Create(ctx, db, storage.EHRS{
		ID:          ksuid.New(),
		Type:        constants.EHRNexHealth,
		Subdomain:   testAgent.EHRSubdomain,
		LocationRef: ptr.To(testAgent.EHRLocationRef),
		LocationID:  locationID,
		CreatedAt:   time.Now(),
	}); err != nil {
		return noop, rerror.Wrap(err)
	}

	return locationID, nil
}

func (t *TestAgent) phoneNumber(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, testAgent constants.TestAgent) (ksuid.KSUID, error) {
	var noop ksuid.KSUID

	existing, err := t.storage.PhoneNumber.Get(ctx, db, storage.PhoneNumberFilters{
		Number:     &testAgent.PhoneNumber,
		PracticeID: &practiceID,
	})
	if err != nil {
		return noop, rerror.Wrap(err)
	}

	if len(existing) > 0 {
		phoneNumber := existing[0]
		phoneNumber.PhoneNumberIDRef = ptr.To(testAgent.PhoneNumberRef)
		phoneNumber.DisabledAt = nil
		phoneNumber.PhoneNumberToActive()
		if err = t.storage.PhoneNumber.Update(ctx, db, phoneNumber); err != nil {
			return noop, rerror.Wrap(err)
		}
		return phoneNumber.ID, nil
	}

	phoneNumberID := ksuid.New()
	if err = t.storage.PhoneNumber.Create(ctx, db, storage.PhoneNumber{
		EntityBase:       storage.EntityBase[states.PhoneNumberState]{EntityState: states.PhoneNumberStateActive},
		ID:               phoneNumberID,
		PhoneNumber:      testAgent.PhoneNumber,
		PhoneNumberIDRef: ptr.To(testAgent.PhoneNumberRef),
		PracticeID:       practiceID,
		CreatedAt:        time.Now(),
	}); err != nil {
		return noop, rerror.Wrap(err)
	}

	return phoneNumberID, nil
}

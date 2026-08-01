package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

func buildInstructions(agentName, practiceName, locationAddress string) string {
	i := strings.NewReplacer(
		"{{AGENT_NAME}}", agentName,
		"{{PRACTICE_NAME}}", practiceName,
		"{{LOCATION_ADDRESS}}", locationAddress,
	)
	return i.Replace(constants.AgentInstructionTemplate)
}

func buildGreeting(agentName, practiceName string) string {
	i := strings.NewReplacer(
		"{{AGENT_NAME}}", agentName,
		"{{PRACTICE_NAME}}", practiceName,
	)
	return i.Replace(constants.AgentGreetingTemplate)
}

type Agent struct {
	*Biz

	storage     storage.Storage
	telnyxAgent agent.IAgent
	landscape   constants.Landscape
}

type IAgent interface {
	Create(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.CreateAgentRequest) (*storage.Agent, error)
	Get(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]storage.Agent, error)
	GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Agent, error)
	Update(ctx context.Context, db storage.DB, id ksuid.KSUID, input contracts.UpdateAgentRequest) (*storage.Agent, error)
	Activate(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Agent, error)
	Delete(ctx context.Context, db storage.DB, id ksuid.KSUID) error
	Disable(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Agent, error)
}

var _ IAgent = (*Agent)(nil)

func (a *Agent) Create(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.CreateAgentRequest) (*storage.Agent, error) {
	if _, err := a.storage.Practice.GetByID(ctx, db, practiceID); err != nil {
		return nil, rerror.Wrap(err)
	}

	if input.LocationID != nil {
		location, err := a.storage.Location.GetByID(ctx, db, *input.LocationID)
		if err != nil {
			return nil, rerror.Wrap(err)
		}
		if location.PracticeID != practiceID {
			return nil, rerror.NewMessage("location not found", rerror.Forbidden)
		}
		if location.EntityState != states.LocationStateActive {
			return nil, rerror.NewMessage("location is not active", rerror.Forbidden)
		}
	}

	if input.PhoneNumberID != nil {
		phoneNumber, err := a.storage.PhoneNumber.GetByID(ctx, db, *input.PhoneNumberID)
		if err != nil {
			return nil, rerror.Wrap(err)
		}
		if phoneNumber.PracticeID != practiceID {
			return nil, rerror.NewMessage("phone number not found", rerror.Forbidden)
		}
		if phoneNumber.EntityState != states.PhoneNumberStateActive {
			return nil, rerror.NewMessage("phone number is not active", rerror.Forbidden)
		}
	}

	agentID := ksuid.New()
	if err := a.storage.Agent.Create(ctx, db, storage.Agent{
		EntityBase: storage.EntityBase[states.AgentState]{
			EntityState: states.AgentStateCreated,
		},
		ID:            agentID,
		PracticeID:    practiceID,
		Name:          input.Name,
		LocationID:    input.LocationID,
		PhoneNumberID: input.PhoneNumberID,
		CreatedAt:     time.Now(),
	}); err != nil {
		return nil, rerror.Wrap(err)
	}

	return a.storage.Agent.GetByID(ctx, db, agentID)
}

func (a *Agent) Update(ctx context.Context, db storage.DB, id ksuid.KSUID, input contracts.UpdateAgentRequest) (*storage.Agent, error) {
	agentRecord, err := a.storage.Agent.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if agentRecord.EntityState == states.AgentStateActive || agentRecord.EntityState == states.AgentStateDisabled {
		return nil, rerror.NewMessage(fmt.Sprintf("cannot update agent in %s state", agentRecord.EntityState), rerror.Forbidden)
	}

	if input.LocationID != nil {
		location, err := a.storage.Location.GetByID(ctx, db, *input.LocationID)
		if err != nil {
			return nil, rerror.Wrap(err)
		}
		if location.PracticeID != agentRecord.PracticeID {
			return nil, rerror.NewMessage("location not found", rerror.Forbidden)
		}
		if location.EntityState != states.LocationStateActive {
			return nil, rerror.NewMessage("location is not active", rerror.Forbidden)
		}

		existing, err := a.storage.Agent.Get(ctx, db, storage.AgentFilters{LocationID: input.LocationID})
		if err != nil {
			return nil, rerror.Wrap(err)
		}
		for _, prev := range existing {
			prev.UnassignLocation()
			if err = a.storage.Agent.Update(ctx, db, prev); err != nil {
				return nil, rerror.Wrap(err)
			}
		}
		agentRecord.AssignLocation(*input.LocationID)
	}

	if input.PhoneNumberID != nil {
		phoneNumber, err := a.storage.PhoneNumber.GetByID(ctx, db, *input.PhoneNumberID)
		if err != nil {
			return nil, rerror.Wrap(err)
		}
		if phoneNumber.PracticeID != agentRecord.PracticeID {
			return nil, rerror.NewMessage("phone number not found", rerror.Forbidden)
		}
		if phoneNumber.EntityState != states.PhoneNumberStateActive {
			return nil, rerror.NewMessage("phone number is not active", rerror.Forbidden)
		}

		existing, err := a.storage.Agent.Get(ctx, db, storage.AgentFilters{PhoneNumberID: input.PhoneNumberID})
		if err != nil {
			return nil, rerror.Wrap(err)
		}
		for _, prev := range existing {
			prev.UnassignPhoneNumber()
			if err = a.storage.Agent.Update(ctx, db, prev); err != nil {
				return nil, rerror.Wrap(err)
			}
		}
		agentRecord.AssignPhoneNumber(*input.PhoneNumberID)
	}

	if input.Name != nil {
		agentRecord.Name = *input.Name
	}

	if input.LocationID != nil || input.PhoneNumberID != nil || input.Name != nil {
		if err = a.storage.Agent.Update(ctx, db, *agentRecord); err != nil {
			return nil, rerror.Wrap(err)
		}
	}

	return a.storage.Agent.GetByID(ctx, db, id)
}

func (a *Agent) Activate(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Agent, error) {
	agentRecord, err := a.storage.Agent.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if agentRecord.AgentRef != nil {
		agentRecord.AgentToActive()
		if err = a.storage.Agent.Update(ctx, db, *agentRecord); err != nil {
			return nil, rerror.Wrap(err)
		}
		return a.storage.Agent.GetByID(ctx, db, id)
	}

	if agentRecord.LocationID == nil {
		return nil, rerror.NewMessage("No location assigned. Please assign a location to activate an agent.", rerror.Forbidden)
	}
	if agentRecord.PhoneNumberID == nil {
		return nil, rerror.NewMessage("agent has no phone number assigned", rerror.Forbidden)
	}

	location, err := a.storage.Location.GetByID(ctx, db, *agentRecord.LocationID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	if location.EntityState != states.LocationStateActive {
		return nil, rerror.NewMessage("assigned location is not active", rerror.Forbidden)
	}

	phoneNumber, err := a.storage.PhoneNumber.GetByID(ctx, db, *agentRecord.PhoneNumberID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	if phoneNumber.EntityState != states.PhoneNumberStateActive {
		return nil, rerror.NewMessage("assigned phone number is not active", rerror.Forbidden)
	}
	if phoneNumber.PhoneNumberIDRef == nil {
		return nil, rerror.NewMessage("assigned phone number has no telnyx reference", rerror.Forbidden)
	}

	practice, err := a.storage.Practice.GetByID(ctx, db, agentRecord.PracticeID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	activeToolState := states.ToolStateActive
	activeTools, err := a.storage.Tool.Get(ctx, db, storage.ToolFilters{EntityState: &activeToolState})
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	toolIDs := make([]string, 0, len(activeTools))
	for _, t := range activeTools {
		if t.ToolRef == "" {
			continue
		}
		toolIDs = append(toolIDs, t.ToolRef)
	}

	result, err := a.telnyxAgent.Create(ctx, agent.CreateAgentParams{
		Name:             agentRecord.Name,
		Model:            constants.DefaultModel,
		Instructions:     buildInstructions(agentRecord.Name, practice.Name, location.Address),
		Greeting:         buildGreeting(agentRecord.Name, practice.Name),
		PhoneNumberIDRef: *phoneNumber.PhoneNumberIDRef,
		ToolIDs:          toolIDs,
	})
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	agentRecord.AgentRef = ptr.To(result.ID)
	agentRecord.Name = result.Name
	agentRecord.AgentToActive()
	if err = a.storage.Agent.Update(ctx, db, *agentRecord); err != nil {
		return nil, rerror.Wrap(err)
	}

	return a.storage.Agent.GetByID(ctx, db, id)
}

func (a *Agent) Get(ctx context.Context, db storage.DB, practiceID ksuid.KSUID) ([]storage.Agent, error) {
	return a.storage.Agent.Get(ctx, db, storage.AgentFilters{PracticeID: &practiceID})
}

func (a *Agent) GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Agent, error) {
	return a.storage.Agent.GetByID(ctx, db, id)
}

func (a *Agent) Disable(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Agent, error) {
	agentRecord, err := a.storage.Agent.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if agentRecord.EntityState != states.AgentStateActive {
		return nil, rerror.NewMessage(fmt.Sprintf("cannot disable agent: not active (state=%s)", agentRecord.EntityState), rerror.Forbidden)
	}

	agentRecord.AgentToDisabled()
	if err = a.storage.Agent.Update(ctx, db, *agentRecord); err != nil {
		return nil, rerror.Wrap(err)
	}

	return a.storage.Agent.GetByID(ctx, db, id)
}

func (a *Agent) Delete(ctx context.Context, db storage.DB, id ksuid.KSUID) error {
	agentRecord, err := a.storage.Agent.GetByID(ctx, db, id)
	if err != nil {
		return rerror.Wrap(err)
	}

	if ptr.From(agentRecord.AgentRef) != constants.TestAgentFor(a.landscape).AgentRef {
		if err = a.telnyxAgent.Delete(ctx, agent.DeleteAgentParams{AgentRef: ptr.From(agentRecord.AgentRef)}); err != nil {
			return rerror.Wrap(err)
		}
	}

	return a.storage.Agent.Delete(ctx, db, id)
}

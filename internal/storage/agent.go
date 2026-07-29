package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type Agent struct {
	EntityBase[states.AgentState]

	ID            ksuid.KSUID  `db:"id" json:"id"`
	PracticeID    ksuid.KSUID  `db:"practice_id" json:"practice_id"`
	AgentRef      *string      `db:"agent_ref" json:"agent_ref"`
	Name          string       `db:"name" json:"name"`
	LocationID    *ksuid.KSUID `db:"location_id" json:"location_id"`
	PhoneNumberID *ksuid.KSUID `db:"phone_number_id" json:"phone_number_id"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     *time.Time   `db:"updated_at" json:"updated_at"`

	Location    *Location    `db:"locations" json:"locations" scan:"notate"`
	PhoneNumber *PhoneNumber `db:"phone_numbers" json:"phone_numbers" scan:"notate"`
}

func (a *Agent) AssignLocation(locationID ksuid.KSUID) {
	a.LocationID = &locationID
	now := time.Now()
	a.UpdatedAt = &now
}

func (a *Agent) UnassignLocation() {
	a.LocationID = nil
	now := time.Now()
	a.UpdatedAt = &now
}

func (a *Agent) AssignPhoneNumber(phoneNumberID ksuid.KSUID) {
	a.PhoneNumberID = &phoneNumberID
	now := time.Now()
	a.UpdatedAt = &now
}

func (a *Agent) UnassignPhoneNumber() {
	a.PhoneNumberID = nil
	now := time.Now()
	a.UpdatedAt = &now
}

func (a *Agent) AgentToActive() {
	a.EntityState = states.AgentStateActive
	now := time.Now()
	a.UpdatedAt = &now
}

func (a *Agent) AgentToDisabled() {
	a.EntityState = states.AgentStateDisabled
	now := time.Now()
	a.UpdatedAt = &now
}

type AgentFilters struct {
	EntityState   *states.AgentState
	PracticeID    *ksuid.KSUID
	LocationID    *ksuid.KSUID
	PhoneNumberID *ksuid.KSUID
	AgentRef      *string
}

type IAgentStorage interface {
	Get(ctx context.Context, db DB, filters AgentFilters) ([]Agent, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Agent, error)
	GetByAgentRef(ctx context.Context, db DB, agentRef string) (*Agent, error)
	GetByPhoneNumberID(ctx context.Context, db DB, phoneNumberID ksuid.KSUID) (*Agent, error)
	Create(ctx context.Context, db DB, agent Agent) error
	Update(ctx context.Context, db DB, agent Agent) error
	Delete(ctx context.Context, db DB, id ksuid.KSUID) error
}

type AgentStorage struct{}

var _ IAgentStorage = (*AgentStorage)(nil)

func (s *AgentStorage) baseSelect() squirrel.SelectBuilder {
	columns := []string{
		"agents.*",
		`0 AS "notate:locations"`,
		"locations.*",
		`0 AS "notate:phone_numbers"`,
		"phone_numbers.*",
	}

	builder := StatementBuilder.
		Select(columns...).
		From("agents").
		LeftJoin("locations ON agents.location_id = locations.id").
		LeftJoin("phone_numbers ON agents.phone_number_id = phone_numbers.id").
		OrderBy("CASE agents.entity_state WHEN 'ACTIVE' THEN 1 WHEN 'CREATED' THEN 2 WHEN 'DISABLED' THEN 3 END", "agents.created_at")

	return builder
}

func (s *AgentStorage) Delete(ctx context.Context, db DB, id ksuid.KSUID) error {
	sql, args, err := StatementBuilder.
		Delete("agents").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *AgentStorage) Get(ctx context.Context, db DB, filters AgentFilters) ([]Agent, error) {
	builder := s.baseSelect()

	if filters.EntityState != nil {
		builder = builder.Where(squirrel.Eq{"agents.entity_state": *filters.EntityState})
	}
	if filters.PracticeID != nil {
		builder = builder.Where(squirrel.Eq{"agents.practice_id": *filters.PracticeID})
	}
	if filters.LocationID != nil {
		builder = builder.Where(squirrel.Eq{"agents.location_id": *filters.LocationID})
	}
	if filters.PhoneNumberID != nil {
		builder = builder.Where(squirrel.Eq{"agents.phone_number_id": *filters.PhoneNumberID})
	}
	if filters.AgentRef != nil {
		builder = builder.Where(squirrel.Eq{"agents.agent_ref": *filters.AgentRef})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	var agents []Agent
	if err = NewScannerRows(ctx, rows).Scan(&agents); err != nil {
		return nil, rerror.New(err)
	}

	return agents, nil
}

func (s *AgentStorage) Create(ctx context.Context, db DB, agent Agent) error {
	sql, args, err := StatementBuilder.Insert("agents").SetMap(map[string]any{
		"entity_state":    agent.EntityState,
		"id":              agent.ID,
		"practice_id":     agent.PracticeID,
		"agent_ref":       agent.AgentRef,
		"name":            agent.Name,
		"location_id":     agent.LocationID,
		"phone_number_id": agent.PhoneNumberID,
		"created_at":      agent.CreatedAt,
	}).ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *AgentStorage) Update(ctx context.Context, db DB, agent Agent) error {
	sql, args, err := StatementBuilder.
		Update("agents").
		SetMap(map[string]any{
			"entity_state":    agent.EntityState,
			"agent_ref":       agent.AgentRef,
			"name":            agent.Name,
			"location_id":     agent.LocationID,
			"phone_number_id": agent.PhoneNumberID,
			"updated_at":      time.Now(),
		}).
		Where(squirrel.Eq{"id": agent.ID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *AgentStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Agent, error) {
	return s.getOne(ctx, db, squirrel.Eq{"agents.id": id})
}

func (s *AgentStorage) GetByAgentRef(ctx context.Context, db DB, agentRef string) (*Agent, error) {
	return s.getOne(ctx, db, squirrel.Eq{"agents.agent_ref": agentRef})
}

func (s *AgentStorage) GetByPhoneNumberID(ctx context.Context, db DB, phoneNumberID ksuid.KSUID) (*Agent, error) {
	return s.getOne(ctx, db, squirrel.Eq{"agents.phone_number_id": phoneNumberID})
}

func (s *AgentStorage) getOne(ctx context.Context, db DB, pred squirrel.Sqlizer) (*Agent, error) {
	sql, args, err := s.baseSelect().Where(pred).ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	var agent Agent
	if err = NewScannerRow(ctx, rows).Scan(&agent); err != nil {
		return nil, rerror.New(err)
	}

	return &agent, nil
}

package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type Conversation struct {
	ID              ksuid.KSUID `db:"id" json:"id"`
	AgentID         ksuid.KSUID `db:"agent_id" json:"agent_id"`
	PhoneNumberID   ksuid.KSUID `db:"phone_number_id" json:"phone_number_id"`
	LocationID      ksuid.KSUID `db:"location_id" json:"location_id"`
	PracticeID      ksuid.KSUID `db:"practice_id" json:"practice_id"`
	ConversationRef string      `db:"conversation_ref" json:"conversation_ref"`
	Duration        int64       `db:"duration" json:"duration"`
	Outcome         string      `db:"outcome" json:"outcome"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
}

type ConversationFilters struct {
	AgentID       *ksuid.KSUID
	PhoneNumberID *ksuid.KSUID
	LocationID    *ksuid.KSUID
	PracticeID    *ksuid.KSUID
}

type IConversationStorage interface {
	Create(ctx context.Context, db DB, conversation Conversation) error
	Get(ctx context.Context, db DB, filters ConversationFilters) ([]Conversation, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Conversation, error)
}

type ConversationStorage struct{}

var _ IConversationStorage = (*ConversationStorage)(nil)

func (s *ConversationStorage) Create(ctx context.Context, db DB, conversation Conversation) error {
	sql, args, err := StatementBuilder.Insert("conversations").SetMap(map[string]any{
		"id":               conversation.ID,
		"agent_id":         conversation.AgentID,
		"phone_number_id":  conversation.PhoneNumberID,
		"location_id":      conversation.LocationID,
		"practice_id":      conversation.PracticeID,
		"conversation_ref": conversation.ConversationRef,
		"duration":         conversation.Duration,
		"outcome":          conversation.Outcome,
		"created_at":       conversation.CreatedAt,
	}).ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *ConversationStorage) Get(ctx context.Context, db DB, filters ConversationFilters) ([]Conversation, error) {
	builder := StatementBuilder.Select("*").From("conversations").OrderBy("created_at DESC")

	if filters.AgentID != nil {
		builder = builder.Where(squirrel.Eq{"agent_id": *filters.AgentID})
	}
	if filters.PhoneNumberID != nil {
		builder = builder.Where(squirrel.Eq{"phone_number_id": *filters.PhoneNumberID})
	}
	if filters.LocationID != nil {
		builder = builder.Where(squirrel.Eq{"location_id": *filters.LocationID})
	}
	if filters.PracticeID != nil {
		builder = builder.Where(squirrel.Eq{"practice_id": *filters.PracticeID})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	conversations, err := pgx.CollectRows(rows, pgx.RowToStructByName[Conversation])
	if err != nil {
		return nil, rerror.New(err)
	}
	return conversations, nil
}

func (s *ConversationStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Conversation, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("conversations").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	conversation, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Conversation])
	if err != nil {
		return nil, rerror.New(err)
	}
	return conversation, nil
}

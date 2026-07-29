package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type Role struct {
	EntityBase[states.RoleState]

	ID         ksuid.KSUID        `db:"id" json:"id"`
	Type       constants.RoleType `db:"type" json:"type"`
	CreatedAt  time.Time          `db:"created_at" json:"created_at"`
	DisabledAt *time.Time         `db:"disabled_at" json:"disabled_at"`
}

type RoleFilters struct {
	Type        *constants.RoleType
	EntityState *states.RoleState
}

type IRoleStorage interface {
	Create(ctx context.Context, db DB, role Role) error
	Get(ctx context.Context, db DB, filters RoleFilters) ([]Role, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Role, error)
	GetByType(ctx context.Context, db DB, roleType constants.RoleType) (*Role, error)
}

type RoleStorage struct{}

var _ IRoleStorage = (*RoleStorage)(nil)

func (s *RoleStorage) Create(ctx context.Context, db DB, role Role) error {
	builder := StatementBuilder.Insert("roles").SetMap(map[string]any{
		"entity_state": role.EntityState,
		"id":           role.ID,
		"type":         role.Type,
		"created_at":   role.CreatedAt,
	})

	sql, args, err := builder.ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *RoleStorage) Get(ctx context.Context, db DB, filters RoleFilters) ([]Role, error) {
	builder := StatementBuilder.Select("*").From("roles")

	if filters.Type != nil {
		builder = builder.Where(squirrel.Eq{"type": *filters.Type})
	}
	if filters.EntityState != nil {
		builder = builder.Where(squirrel.Eq{"entity_state": *filters.EntityState})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	roles, err := pgx.CollectRows(rows, pgx.RowToStructByName[Role])
	if err != nil {
		return nil, rerror.New(err)
	}
	return roles, nil
}

func (s *RoleStorage) GetByType(ctx context.Context, db DB, roleType constants.RoleType) (*Role, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("roles").
		Where(squirrel.Eq{"type": roleType}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	role, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Role])
	if err != nil {
		return nil, rerror.New(err)
	}
	return role, nil
}

func (s *RoleStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Role, error) {
	builder := StatementBuilder.
		Select("*").
		From("roles").
		Where(squirrel.Eq{"id": id})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	role, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Role])
	if err != nil {
		return nil, rerror.New(err)
	}
	return role, nil
}

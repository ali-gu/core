package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type Tool struct {
	EntityBase[states.ToolState]

	ID         ksuid.KSUID        `db:"id" json:"id"`
	Type       constants.ToolType `db:"type" json:"type"`
	Kind       constants.ToolKind `db:"kind" json:"kind"`
	ToolRef    string             `db:"tool_ref" json:"tool_ref"`
	Config     map[string]any     `db:"config" json:"config"`
	CreatedAt  time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt  *time.Time         `db:"updated_at" json:"updated_at"`
	DisabledAt *time.Time         `db:"disabled_at" json:"disabled_at"`
}

func (t *Tool) ToolToDisabled() {
	t.EntityState = states.ToolStateDisabled
	t.DisabledAt = ptr.To(time.Now())
	t.UpdatedAt = ptr.To(time.Now())
}

type ToolFilters struct {
	EntityState *states.ToolState
	Type        *constants.ToolType
	Kind        *constants.ToolKind
}

type IToolStorage interface {
	Create(ctx context.Context, db DB, tool Tool) error
	Get(ctx context.Context, db DB, filters ToolFilters) ([]Tool, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Tool, error)
	GetByKind(ctx context.Context, db DB, kind constants.ToolKind) (*Tool, error)
	Update(ctx context.Context, db DB, tool Tool) error
}

type ToolStorage struct{}

var _ IToolStorage = (*ToolStorage)(nil)

func configExpr(config map[string]any) (squirrel.Sqlizer, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, rerror.New(err)
	}
	return squirrel.Expr("?::jsonb", string(raw)), nil
}

func (s *ToolStorage) Create(ctx context.Context, db DB, tool Tool) error {
	config, err := configExpr(tool.Config)
	if err != nil {
		return err
	}

	sql, args, err := StatementBuilder.Insert("tools").SetMap(map[string]any{
		"entity_state": tool.EntityState,
		"id":           tool.ID,
		"type":         tool.Type,
		"kind":         tool.Kind,
		"tool_ref":     tool.ToolRef,
		"config":       config,
		"created_at":   tool.CreatedAt,
	}).ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *ToolStorage) Get(ctx context.Context, db DB, filters ToolFilters) ([]Tool, error) {
	builder := StatementBuilder.Select("*").From("tools")

	if filters.EntityState != nil {
		builder = builder.Where(squirrel.Eq{"entity_state": *filters.EntityState})
	}
	if filters.Type != nil {
		builder = builder.Where(squirrel.Eq{"type": *filters.Type})
	}
	if filters.Kind != nil {
		builder = builder.Where(squirrel.Eq{"kind": *filters.Kind})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	tools, err := pgx.CollectRows(rows, pgx.RowToStructByName[Tool])
	if err != nil {
		return nil, rerror.New(err)
	}
	return tools, nil
}

func (s *ToolStorage) GetByKind(ctx context.Context, db DB, kind constants.ToolKind) (*Tool, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("tools").
		Where(squirrel.Eq{"kind": kind}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	tool, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Tool])
	if err != nil {
		return nil, rerror.New(err)
	}
	return tool, nil
}

func (s *ToolStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Tool, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("tools").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	tool, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Tool])
	if err != nil {
		return nil, rerror.New(err)
	}
	return tool, nil
}

func (s *ToolStorage) Update(ctx context.Context, db DB, tool Tool) error {
	config, err := configExpr(tool.Config)
	if err != nil {
		return err
	}

	sql, args, err := StatementBuilder.
		Update("tools").
		SetMap(map[string]any{
			"entity_state": tool.EntityState,
			"tool_ref":     tool.ToolRef,
			"config":       config,
			"disabled_at":  tool.DisabledAt,
			"updated_at":   time.Now(),
		}).
		Where(squirrel.Eq{"id": tool.ID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

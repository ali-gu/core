package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type Practice struct {
	EntityBase[states.PracticeState]

	ID         ksuid.KSUID `db:"id" json:"id"`
	Name       string      `db:"name" json:"name"`
	Email      *string     `db:"email" json:"email"`
	ZipCode    *string     `db:"zip_code" json:"zip_code"`
	Website    *string     `db:"website" json:"website"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt  *time.Time  `db:"updated_at" json:"updated_at"`
	DisabledAt *time.Time  `db:"disabled_at" json:"disabled_at"`
}

type IPracticeStorage interface {
	Create(ctx context.Context, db DB, practice Practice) error
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Practice, error)
	Update(ctx context.Context, db DB, practice Practice) error
}

type PracticeStorage struct{}

var _ IPracticeStorage = (*PracticeStorage)(nil)

func (s *PracticeStorage) Create(ctx context.Context, db DB, practice Practice) error {
	builder := StatementBuilder.Insert("practices").SetMap(map[string]any{
		"entity_state": practice.EntityState,
		"id":           practice.ID,
		"name":         practice.Name,
		"email":        practice.Email,
		"zip_code":     practice.ZipCode,
		"website":      practice.Website,
		"created_at":   practice.CreatedAt,
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

func (s *PracticeStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Practice, error) {
	builder := StatementBuilder.
		Select("*").
		From("practices").
		Where(squirrel.Eq{"id": id})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	practice, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[Practice])
	if err != nil {
		return nil, rerror.New(err)
	}
	return practice, nil
}

func (s *PracticeStorage) Update(ctx context.Context, db DB, practice Practice) error {
	sql, args, err := StatementBuilder.
		Update("practices").
		SetMap(map[string]any{
			"email":      practice.Email,
			"zip_code":   practice.ZipCode,
			"website":    practice.Website,
			"updated_at": time.Now(),
		}).
		Where(squirrel.Eq{"id": practice.ID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

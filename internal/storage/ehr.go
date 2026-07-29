package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type EHRS struct {
	ID            ksuid.KSUID   `db:"id" json:"id"`
	Type          constants.EHR `db:"type" json:"type"`
	Subdomain     string        `db:"subdomain" json:"subdomain"`
	LocationRef   *string       `db:"location_ref" json:"location_ref"`
	LocationID    ksuid.KSUID   `db:"location_id" json:"location_id"`
	OnboardingURL string        `db:"onboarding_url" json:"onboarding_url"`
	OnboardingID  string        `db:"onboarding_id" json:"onboarding_id"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
}

type EHRFilters struct {
	LocationID *ksuid.KSUID
}

type IEHRStorage interface {
	Create(ctx context.Context, db DB, ehr EHRS) error
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*EHRS, error)
	GetByLocationID(ctx context.Context, db DB, locationID ksuid.KSUID) (*EHRS, error)
	Get(ctx context.Context, db DB, filters EHRFilters) ([]EHRS, error)
}

type EHRStorage struct{}

var _ IEHRStorage = (*EHRStorage)(nil)

func (s *EHRStorage) Create(ctx context.Context, db DB, ehr EHRS) error {
	builder := StatementBuilder.Insert("ehrs").SetMap(map[string]any{
		"id":             ehr.ID,
		"type":           ehr.Type,
		"subdomain":      ehr.Subdomain,
		"location_ref":   ehr.LocationRef,
		"location_id":    ehr.LocationID,
		"onboarding_url": ehr.OnboardingURL,
		"onboarding_id":  ehr.OnboardingID,
		"created_at":     ehr.CreatedAt,
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

func (s *EHRStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*EHRS, error) {
	builder := StatementBuilder.
		Select("*").
		From("ehrs").
		Where(squirrel.Eq{"id": id})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	ehr, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[EHRS])
	if err != nil {
		return nil, rerror.New(err)
	}
	return ehr, nil
}

func (s *EHRStorage) GetByLocationID(ctx context.Context, db DB, locationID ksuid.KSUID) (*EHRS, error) {
	builder := StatementBuilder.
		Select("*").
		From("ehrs").
		Where(squirrel.Eq{"location_id": locationID})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	ehr, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[EHRS])
	if err != nil {
		return nil, rerror.New(err)
	}
	return ehr, nil
}

func (s *EHRStorage) Get(ctx context.Context, db DB, filters EHRFilters) ([]EHRS, error) {
	builder := StatementBuilder.Select("*").From("ehrs")

	if filters.LocationID != nil {
		builder = builder.Where(squirrel.Eq{"location_id": *filters.LocationID})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	ehrs, err := pgx.CollectRows(rows, pgx.RowToStructByName[EHRS])
	if err != nil {
		return nil, rerror.New(err)
	}
	return ehrs, nil
}

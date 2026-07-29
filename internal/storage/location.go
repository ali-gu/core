package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type Location struct {
	EntityBase[states.LocationState]

	ID         ksuid.KSUID `db:"id" json:"id"`
	Address    string      `db:"address" json:"address"`
	PracticeID ksuid.KSUID `db:"practice_id" json:"practice_id"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt  *time.Time  `db:"updated_at" json:"updated_at"`
	DisabledAt *time.Time  `db:"disabled_at" json:"disabled_at"`

	EHR *EHRS `db:"ehrs" json:"ehrs" scan:"notate"`
}

func (l *Location) LocationToActive() {
	l.EntityState = states.LocationStateActive
	l.UpdatedAt = ptr.To(time.Now())
}

func (l *Location) LocationToDisabled() {
	l.EntityState = states.LocationStateDisabled
	l.DisabledAt = ptr.To(time.Now())
}

type LocationFilters struct {
	PracticeID  *ksuid.KSUID
	EntityState *states.LocationState
}

type ILocationStorage interface {
	Create(ctx context.Context, db DB, location Location) error
	Get(ctx context.Context, db DB, filters LocationFilters) ([]Location, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Location, error)
	Update(ctx context.Context, db DB, location Location) error
	Delete(ctx context.Context, db DB, id ksuid.KSUID) error
}

type LocationStorage struct{}

var _ ILocationStorage = (*LocationStorage)(nil)

func (s *LocationStorage) Create(ctx context.Context, db DB, location Location) error {
	sql, args, err := StatementBuilder.Insert("locations").SetMap(map[string]any{
		"entity_state": location.EntityState,
		"id":           location.ID,
		"address":      location.Address,
		"practice_id":  location.PracticeID,
		"created_at":   location.CreatedAt,
	}).ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *LocationStorage) baseSelect() squirrel.SelectBuilder {
	columns := []string{
		"locations.*",
		`0 AS "notate:ehrs"`,
		"ehrs.*",
	}

	return StatementBuilder.
		Select(columns...).
		From("locations").
		LeftJoin("ehrs ON ehrs.location_id = locations.id").
		OrderBy("CASE locations.entity_state WHEN 'ACTIVE' THEN 1 WHEN 'PENDING' THEN 2 WHEN 'DISABLED' THEN 3 END")
}

func (s *LocationStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*Location, error) {
	sql, args, err := s.baseSelect().
		Where(squirrel.Eq{"locations.id": id}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	var location Location
	if err = NewScannerRow(ctx, rows).Scan(&location); err != nil {
		return nil, rerror.New(err)
	}

	return &location, nil
}

func (s *LocationStorage) Get(ctx context.Context, db DB, filters LocationFilters) ([]Location, error) {
	builder := s.baseSelect()

	if filters.PracticeID != nil {
		builder = builder.Where(squirrel.Eq{"locations.practice_id": *filters.PracticeID})
	}
	if filters.EntityState != nil {
		builder = builder.Where(squirrel.Eq{"locations.entity_state": *filters.EntityState})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	var locations []Location
	if err = NewScannerRows(ctx, rows).Scan(&locations); err != nil {
		return nil, rerror.New(err)
	}

	return locations, nil
}

func (s *LocationStorage) Delete(ctx context.Context, db DB, id ksuid.KSUID) error {
	sql, args, err := StatementBuilder.
		Delete("locations").
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

func (s *LocationStorage) Update(ctx context.Context, db DB, location Location) error {
	sql, args, err := StatementBuilder.
		Update("locations").
		SetMap(map[string]any{
			"entity_state": location.EntityState,
			"disabled_at":  location.DisabledAt,
			"updated_at":   time.Now(),
		}).
		Where(squirrel.Eq{"id": location.ID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

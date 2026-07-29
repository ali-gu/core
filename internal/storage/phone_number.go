package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type PhoneNumber struct {
	EntityBase[states.PhoneNumberState]

	ID                        ksuid.KSUID `db:"id" json:"id"`
	PhoneNumber               string      `db:"number" json:"number"`
	PhoneNumberIDRef          *string     `db:"phone_number_ref" json:"phone_number_ref"`
	PhoneNumberReservationRef *string     `db:"phone_number_reservation_ref" json:"phone_number_reservation_ref"`
	PracticeID                ksuid.KSUID `db:"practice_id" json:"practice_id"`
	CreatedAt                 time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt                 *time.Time  `db:"updated_at" json:"updated_at"`
	DisabledAt                *time.Time  `db:"disabled_at" json:"disabled_at"`
}

func (p *PhoneNumber) PhoneNumberToActive() {
	p.EntityState = states.PhoneNumberStateActive
	p.UpdatedAt = ptr.To(time.Now())
}

func (p *PhoneNumber) PhoneNumberToDisabled() {
	p.EntityState = states.PhoneNumberStateDisabled
	p.DisabledAt = ptr.To(time.Now())
}

type PhoneNumberFilters struct {
	Number      *string
	PracticeID  *ksuid.KSUID
	EntityState *states.PhoneNumberState
}

type IPhoneNumberStorage interface {
	Create(ctx context.Context, db DB, phoneNumber PhoneNumber) error
	Get(ctx context.Context, db DB, filters PhoneNumberFilters) ([]PhoneNumber, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*PhoneNumber, error)
	GetByNumber(ctx context.Context, db DB, number string) (*PhoneNumber, error)
	Update(ctx context.Context, db DB, phoneNumber PhoneNumber) error
	Delete(ctx context.Context, db DB, id ksuid.KSUID) error
}

type PhoneNumberStorage struct{}

var _ IPhoneNumberStorage = (*PhoneNumberStorage)(nil)

func (s *PhoneNumberStorage) Create(ctx context.Context, db DB, phoneNumber PhoneNumber) error {
	sql, args, err := StatementBuilder.
		Insert("phone_numbers").
		SetMap(map[string]any{
			"entity_state":                 phoneNumber.EntityState,
			"id":                           phoneNumber.ID,
			"number":                       phoneNumber.PhoneNumber,
			"phone_number_ref":             phoneNumber.PhoneNumberIDRef,
			"phone_number_reservation_ref": phoneNumber.PhoneNumberReservationRef,
			"practice_id":                  phoneNumber.PracticeID,
			"created_at":                   phoneNumber.CreatedAt,
		}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *PhoneNumberStorage) Get(ctx context.Context, db DB, filters PhoneNumberFilters) ([]PhoneNumber, error) {
	builder := StatementBuilder.Select("*").From("phone_numbers")

	if filters.Number != nil {
		builder = builder.Where(squirrel.Eq{"number": *filters.Number})
	}
	if filters.PracticeID != nil {
		builder = builder.Where(squirrel.Eq{"practice_id": *filters.PracticeID})
	}
	if filters.EntityState != nil {
		builder = builder.Where(squirrel.Eq{"entity_state": *filters.EntityState})
	}

	builder = builder.OrderBy("CASE entity_state WHEN 'ACTIVE' THEN 1 WHEN 'RESERVED' THEN 2 END")

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	phoneNumbers, err := pgx.CollectRows(rows, pgx.RowToStructByName[PhoneNumber])
	if err != nil {
		return nil, rerror.New(err)
	}
	return phoneNumbers, nil
}

func (s *PhoneNumberStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*PhoneNumber, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("phone_numbers").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	phoneNumber, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[PhoneNumber])
	if err != nil {
		return nil, rerror.New(err)
	}
	return phoneNumber, nil
}

func (s *PhoneNumberStorage) GetByNumber(ctx context.Context, db DB, number string) (*PhoneNumber, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("phone_numbers").
		Where(squirrel.Eq{"number": number}).
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	phoneNumber, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[PhoneNumber])
	if err != nil {
		return nil, rerror.New(err)
	}
	return phoneNumber, nil
}

func (s *PhoneNumberStorage) Update(ctx context.Context, db DB, phoneNumber PhoneNumber) error {
	sql, args, err := StatementBuilder.
		Update("phone_numbers").
		SetMap(map[string]any{
			"entity_state":     phoneNumber.EntityState,
			"phone_number_ref": phoneNumber.PhoneNumberIDRef,
			"disabled_at":      phoneNumber.DisabledAt,
			"updated_at":       time.Now(),
		}).
		Where(squirrel.Eq{"id": phoneNumber.ID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *PhoneNumberStorage) Delete(ctx context.Context, db DB, id ksuid.KSUID) error {
	sql, args, err := StatementBuilder.
		Delete("phone_numbers").
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

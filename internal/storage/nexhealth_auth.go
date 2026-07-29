package storage

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
)

const nexHealthAuthTokenID = "singleton"

type NexHealthAuthToken struct {
	ID        string     `db:"id" json:"id"`
	Token     string     `db:"token" json:"token"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}

type INexHealthAuthStorage interface {
	LockForUpdate(ctx context.Context, db DB) (*NexHealthAuthToken, error)
	Update(ctx context.Context, db DB, token string, expiresAt time.Time) error
}

type NexHealthAuthStorage struct{}

var _ INexHealthAuthStorage = (*NexHealthAuthStorage)(nil)

func (s *NexHealthAuthStorage) LockForUpdate(ctx context.Context, db DB) (*NexHealthAuthToken, error) {
	sql, args, err := StatementBuilder.
		Select("*").
		From("nexhealth_auth_tokens").
		Where(squirrel.Eq{"id": nexHealthAuthTokenID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	token, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[NexHealthAuthToken])
	if err != nil {
		return nil, rerror.New(err)
	}
	return token, nil
}

func (s *NexHealthAuthStorage) Update(ctx context.Context, db DB, token string, expiresAt time.Time) error {
	now := time.Now()
	sql, args, err := StatementBuilder.
		Update("nexhealth_auth_tokens").
		Set("token", token).
		Set("expires_at", expiresAt).
		Set("updated_at", now).
		Where(squirrel.Eq{"id": nexHealthAuthTokenID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

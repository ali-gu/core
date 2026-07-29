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

type User struct {
	EntityBase[states.UserState]

	ID         ksuid.KSUID `db:"id" json:"id"`
	UserRef    string      `db:"user_ref" json:"user_ref"`
	RoleID     ksuid.KSUID `db:"role_id" json:"role_id"`
	PracticeID ksuid.KSUID `db:"practice_id" json:"practice_id"`
	Email      string      `db:"email" json:"email"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt  *time.Time  `db:"updated_at" json:"updated_at"`
	DisabledAt *time.Time  `db:"disabled_at" json:"disabled_at"`

	Practice Practice `db:"practices" json:"practices" scan:"notate"`
	Role     Role     `db:"roles" json:"roles" scan:"notate"`
}

func (u *User) IsInvited() bool {
	return u.EntityState == states.UserStateInvited
}

func (u *User) IsDisabled() bool {
	return u.EntityState == states.UserStateDisabled
}

func (u *User) IsActive() bool {
	return u.EntityState == states.UserStateActive
}

func (u *User) InvitationToActive(supabaseUserID string, roleID ksuid.KSUID) {
	u.EntityState = states.UserStateActive
	u.UserRef = supabaseUserID
	u.RoleID = roleID
	u.UpdatedAt = ptr.To(time.Now())
}

type UserFilters struct {
	Email       *string
	UserRef     *string
	EntityState *states.UserState
}

type IUserStorage interface {
	Create(ctx context.Context, db DB, user User) error
	Get(ctx context.Context, db DB, filters UserFilters) ([]User, error)
	GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*User, error)
	GetByEmail(ctx context.Context, db DB, email string) (*User, error)
	GetByUserRef(ctx context.Context, db DB, userRef string) (*User, error)
	Update(ctx context.Context, db DB, user User) error
}

type UserStorage struct{}

var _ IUserStorage = (*UserStorage)(nil)

func (s *UserStorage) baseSelect() squirrel.SelectBuilder {
	columns := []string{
		"users.*",
		`0 AS "notate:practices"`,
		"practices.*",
		`0 AS "notate:roles"`,
		"roles.*",
	}

	builder := StatementBuilder.
		Select(columns...).
		From("users").
		Join("practices ON practices.id = users.practice_id").
		Join("roles ON roles.id = users.role_id")

	return builder
}

func (s *UserStorage) Create(ctx context.Context, db DB, user User) error {
	sql, args, err := StatementBuilder.Insert("users").SetMap(map[string]any{
		"entity_state": user.EntityState,
		"id":           user.ID,
		"user_ref":     user.UserRef,
		"role_id":      user.RoleID,
		"practice_id":  user.PracticeID,
		"email":        user.Email,
		"created_at":   user.CreatedAt,
	}).ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

func (s *UserStorage) GetByID(ctx context.Context, db DB, id ksuid.KSUID) (*User, error) {
	return s.getOne(ctx, db, squirrel.Eq{"users.id": id})
}

func (s *UserStorage) GetByEmail(ctx context.Context, db DB, email string) (*User, error) {
	return s.getOne(ctx, db, squirrel.Eq{"users.email": email})
}

func (s *UserStorage) GetByUserRef(ctx context.Context, db DB, userRef string) (*User, error) {
	return s.getOne(ctx, db, squirrel.Eq{"users.user_ref": userRef})
}

func (s *UserStorage) getOne(ctx context.Context, db DB, pred squirrel.Sqlizer) (*User, error) {
	sql, args, err := s.baseSelect().Where(pred).ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	var user User
	if err = NewScannerRow(ctx, rows).Scan(&user); err != nil {
		return nil, rerror.New(err)
	}

	return &user, nil
}

func (s *UserStorage) Get(ctx context.Context, db DB, filters UserFilters) ([]User, error) {
	builder := s.baseSelect()

	if filters.Email != nil {
		builder = builder.Where(squirrel.Eq{"users.email": *filters.Email})
	}
	if filters.UserRef != nil {
		builder = builder.Where(squirrel.Eq{"users.user_ref": *filters.UserRef})
	}
	if filters.EntityState != nil {
		builder = builder.Where(squirrel.Eq{"users.entity_state": *filters.EntityState})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, rerror.New(err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, rerror.New(err)
	}

	var users []User
	if err = NewScannerRows(ctx, rows).Scan(&users); err != nil {
		return nil, rerror.New(err)
	}

	return users, nil
}

func (s *UserStorage) Update(ctx context.Context, db DB, user User) error {
	sql, args, err := StatementBuilder.
		Update("users").
		SetMap(map[string]any{
			"entity_state": user.EntityState,
			"user_ref":     user.UserRef,
			"role_id":      user.RoleID,
		}).
		Where(squirrel.Eq{"id": user.ID}).
		ToSql()
	if err != nil {
		return rerror.New(err)
	}

	if _, err = db.Exec(ctx, sql, args...); err != nil {
		return rerror.New(err)
	}
	return nil
}

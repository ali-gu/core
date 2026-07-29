package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type User struct {
	*Biz

	storage      storage.Storage
	supabaseAuth auth.IAuth
}

type IUser interface {
	SignUp(ctx context.Context, db storage.DB, input contracts.SignUpRequest) (*storage.User, error)

	SignIn(ctx context.Context, db storage.DB, input contracts.SignInRequest) (*storage.User, *auth.Session, error)
	Refresh(ctx context.Context, db storage.DB, refreshToken string) (*storage.User, *auth.Session, error)
	Authenticate(ctx context.Context, db storage.DB, accessToken string) (*storage.User, error)
	SignOut(ctx context.Context, db storage.DB, accessToken string) error
	Invite(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.InviteUserRequest) (*storage.User, error)
}

var _ IUser = (*User)(nil)

func (u *User) SignUp(ctx context.Context, db storage.DB, input contracts.SignUpRequest) (*storage.User, error) {
	existingUser, err := u.storage.User.GetByEmail(ctx, db, input.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, rerror.Wrap(err)
	}

	if existingUser != nil && !existingUser.IsInvited() {
		return nil, rerror.NewMessage(fmt.Sprintf("a user with email %q already exists", input.Email), rerror.Validation)
	}

	if existingUser == nil && input.PracticeName == nil {
		return nil, rerror.NewMessage("practice name is required for a new user", rerror.Validation)
	}

	supabaseUser, err := u.supabaseAuth.SignUp(ctx, input.Email, input.Password)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if existingUser != nil {
		role, err := u.roleByType(ctx, db, constants.RoleTypeReader)
		if err != nil {
			return nil, rerror.Wrap(err)
		}

		existingUser.InvitationToActive(supabaseUser.ID, role.ID)
		if err := u.storage.User.Update(ctx, db, *existingUser); err != nil {
			return nil, rerror.Wrap(err)
		}

		return u.storage.User.GetByID(ctx, db, existingUser.ID)
	}

	if input.PracticeName == nil {
		return nil, rerror.NewMessage("practice name is required to sign up a new user", rerror.Validation)
	}

	role, err := u.roleByType(ctx, db, constants.RoleTypeAdmin)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	practice, err := u.Biz.Practice.Create(ctx, db, contracts.CreatePracticeRequest{Name: *input.PracticeName})
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	userID := ksuid.New()
	if err := u.storage.User.Create(ctx, db, storage.User{
		EntityBase: storage.EntityBase[states.UserState]{EntityState: states.UserStateActive},
		ID:         userID,
		UserRef:    supabaseUser.ID,
		RoleID:     role.ID,
		PracticeID: practice.ID,
		Email:      input.Email,
		CreatedAt:  time.Now(),
	}); err != nil {
		return nil, rerror.Wrap(err)
	}

	return u.storage.User.GetByID(ctx, db, userID)
}

func (u *User) SignIn(ctx context.Context, db storage.DB, input contracts.SignInRequest) (*storage.User, *auth.Session, error) {
	result, err := u.supabaseAuth.SignIn(ctx, input.Email, input.Password)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	user, err := u.activeUserByRef(ctx, db, result.ID)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	return user, &result.Session, nil
}

func (u *User) Refresh(ctx context.Context, db storage.DB, refreshToken string) (*storage.User, *auth.Session, error) {
	result, err := u.supabaseAuth.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	user, err := u.activeUserByRef(ctx, db, result.ID)
	if err != nil {
		return nil, nil, rerror.Wrap(err)
	}

	return user, &result.Session, nil
}

func (u *User) SignOut(ctx context.Context, _ storage.DB, accessToken string) error {
	return rerror.Wrap(u.supabaseAuth.SignOut(ctx, accessToken))
}

func (u *User) Authenticate(ctx context.Context, db storage.DB, accessToken string) (*storage.User, error) {
	identity, err := u.supabaseAuth.Authenticate(ctx, accessToken)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	return u.activeUserByRef(ctx, db, identity.ID)
}

func (u *User) activeUserByRef(ctx context.Context, db storage.DB, ref string) (*storage.User, error) {
	user, err := u.storage.User.GetByUserRef(ctx, db, ref)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if !user.IsActive() {
		return nil, rerror.NewMessage("user account is not active", rerror.Forbidden)
	}

	return user, nil
}

func (u *User) Invite(ctx context.Context, db storage.DB, practiceID ksuid.KSUID, input contracts.InviteUserRequest) (*storage.User, error) {
	existing, err := u.storage.User.GetByEmail(ctx, db, input.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, rerror.Wrap(err)
	}
	if existing != nil {
		if existing.PracticeID == practiceID {
			return nil, rerror.NewMessage(fmt.Sprintf("a user with email %q already exists in the system", input.Email), rerror.Validation)
		}
		return nil, rerror.NewMessage(fmt.Sprintf("user with email %q is associated with another practice and must be removed from that practice first", input.Email), rerror.Validation)
	}

	role, err := u.roleByType(ctx, db, constants.RoleTypeReader)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	invited, err := u.supabaseAuth.Invite(ctx, input.Email)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	userID := ksuid.New()
	if err := u.storage.User.Create(ctx, db, storage.User{
		EntityBase: storage.EntityBase[states.UserState]{EntityState: states.UserStateInvited},
		ID:         userID,
		UserRef:    invited.ID,
		RoleID:     role.ID,
		PracticeID: practiceID,
		Email:      input.Email,
		CreatedAt:  time.Now(),
	}); err != nil {
		return nil, rerror.Wrap(err)
	}

	return u.storage.User.GetByID(ctx, db, userID)
}

func (u *User) roleByType(ctx context.Context, db storage.DB, roleType constants.RoleType) (*storage.Role, error) {
	role, err := u.storage.Role.GetByType(ctx, db, roleType)
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	if role.EntityState != states.RoleStateActive {
		return nil, rerror.New(fmt.Errorf("biz: no active role found for type %q", roleType))
	}

	return role, nil
}

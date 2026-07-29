package biz_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func seedRole(t *testing.T, cfg *testutils.TestConfig, roleType constants.RoleType) storage.Role {
	t.Helper()

	roles, err := cfg.Deps.Storage.Role.Get(cfg.Ctx, cfg.DB, storage.RoleFilters{Type: &roleType})
	require.NoError(t, err)
	require.Len(t, roles, 1)
	return roles[0]
}

func clearRoleType(t *testing.T, cfg *testutils.TestConfig, roleType constants.RoleType) {
	t.Helper()

	_, err := cfg.DB.Exec(cfg.Ctx, "DELETE FROM roles WHERE type = $1", roleType)
	require.NoError(t, err)
}

func seedUser(t *testing.T, cfg *testutils.TestConfig, email string, state states.UserState, roleType constants.RoleType) storage.User {
	t.Helper()

	role := seedRole(t, cfg, roleType)

	practiceID := ksuid.New()
	require.NoError(t, cfg.Deps.Storage.Practice.Create(cfg.Ctx, cfg.DB, storage.Practice{
		EntityBase: storage.EntityBase[states.PracticeState]{EntityState: states.PracticeStateActive},
		ID:         practiceID,
		Name:       "seeded_practice",
		CreatedAt:  time.Now(),
	}))

	userID := ksuid.New()
	require.NoError(t, cfg.Deps.Storage.User.Create(cfg.Ctx, cfg.DB, storage.User{
		EntityBase: storage.EntityBase[states.UserState]{EntityState: state},
		ID:         userID,
		UserRef:    "seeded_supabase_ref",
		RoleID:     role.ID,
		PracticeID: practiceID,
		Email:      email,
		CreatedAt:  time.Now(),
	}))

	user, err := cfg.Deps.Storage.User.GetByID(cfg.Ctx, cfg.DB, userID)
	require.NoError(t, err)
	return *user
}

func newUserBizWithAuthMock(t *testing.T, authMock *auth.MockIAuth) (*testutils.TestConfig, biz.Biz) {
	t.Helper()

	return testutils.BasicSetupWithDeps(t, biz.Dependencies{
		Storage:                  storage.NewStorage(),
		TelnyxAgent:              agent.NewMockIAgent(t),
		TelnyxPhoneNumberManager: phonenumber.NewMockIPhoneNumberManager(t),
		SupabaseAuth:             authMock,
		EHR:                      ehr.NewMockIEHR(t),
	})
}

func Test_User_SignUp(t *testing.T) {
	t.Run("success_creates_a_new_practice_and_uses_the_existing_admin_role_when_user_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		adminRole := seedRole(t, cfg, constants.RoleTypeAdmin)

		user, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "new_user@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("new_user_practice"),
		})
		require.NoError(t, err)
		require.NotEmpty(t, user.ID)
		require.Equal(t, "new_user@example.com", user.Email)
		require.Equal(t, states.UserStateActive, user.EntityState)
		require.NotEmpty(t, user.PracticeID)
		require.Equal(t, adminRole.ID, user.RoleID)
		require.Equal(t, "supabase_user_ref", user.UserRef)

		practice, err := cfg.Deps.Storage.Practice.GetByID(cfg.Ctx, cfg.DB, user.PracticeID)
		require.NoError(t, err)
		require.Equal(t, user.PracticeID, practice.ID)
		require.Equal(t, "new_user_practice", practice.Name)
	})

	t.Run("error_when_practice_name_is_missing_for_a_new_user", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		seedRole(t, cfg, constants.RoleTypeAdmin)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "no_practice_name@example.com",
			Password: "S3cr3t-pw",
		})
		require.Error(t, err)

		email := "no_practice_name@example.com"
		users, err := cfg.Deps.Storage.User.Get(cfg.Ctx, cfg.DB, storage.UserFilters{Email: &email})
		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("success_activates_an_invited_user_and_uses_the_existing_reader_role", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		readerRole := seedRole(t, cfg, constants.RoleTypeReader)

		invited := seedUser(t, cfg, "invited_user@example.com", states.UserStateInvited, constants.RoleTypeAdmin)

		user, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "invited_user@example.com",
			Password: "S3cr3t-pw",
		})
		require.NoError(t, err)
		require.Equal(t, invited.ID, user.ID)
		require.Equal(t, invited.PracticeID, user.PracticeID)
		require.Equal(t, states.UserStateActive, user.EntityState)
		require.Equal(t, readerRole.ID, user.RoleID)
		require.Equal(t, "supabase_user_ref", user.UserRef)
	})

	t.Run("error_when_no_active_admin_role_exists", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		clearRoleType(t, cfg, constants.RoleTypeAdmin)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "no_admin_role@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("no_admin_role_practice"),
		})
		require.Error(t, err)

		email := "no_admin_role@example.com"
		users, err := cfg.Deps.Storage.User.Get(cfg.Ctx, cfg.DB, storage.UserFilters{Email: &email})
		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("error_when_no_active_reader_role_exists_leaves_invited_user_untouched", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		invited := seedUser(t, cfg, "no_reader_role@example.com", states.UserStateInvited, constants.RoleTypeAdmin)
		clearRoleType(t, cfg, constants.RoleTypeReader)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "no_reader_role@example.com",
			Password: "S3cr3t-pw",
		})
		require.Error(t, err)

		unchanged, err := cfg.Deps.Storage.User.GetByID(cfg.Ctx, cfg.DB, invited.ID)
		require.NoError(t, err)
		require.Equal(t, states.UserStateInvited, unchanged.EntityState)
		require.Equal(t, invited.RoleID, unchanged.RoleID)
	})

	t.Run("error_when_user_already_active", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		seedUser(t, cfg, "active_user@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "active_user@example.com",
			Password: "S3cr3t-pw",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("error_when_user_already_disabled", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		seedUser(t, cfg, "disabled_user@example.com", states.UserStateDisabled, constants.RoleTypeAdmin)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "disabled_user@example.com",
			Password: "S3cr3t-pw",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("error_when_signing_up_twice_with_the_same_email", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		seedRole(t, cfg, constants.RoleTypeAdmin)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "dup_user@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("dup_user_practice"),
		})
		require.NoError(t, err)

		_, err = bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "dup_user@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("dup_user_practice"),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("success_calls_supabase_signup_with_the_given_credentials", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		seedRole(t, cfg, constants.RoleTypeAdmin)
		authMock := cfg.Deps.SupabaseAuth.(*auth.MockIAuth)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "checked_user@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("checked_user_practice"),
		})
		require.NoError(t, err)
		authMock.AssertCalled(t, "SignUp", mock.Anything, "checked_user@example.com", "S3cr3t-pw")
	})

	t.Run("error_when_supabase_signup_fails_leaves_no_user_created", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignUp", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("supabase unavailable")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "fail_user@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("fail_user_practice"),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "supabase unavailable")

		email := "fail_user@example.com"
		users, err := cfg.Deps.Storage.User.Get(cfg.Ctx, cfg.DB, storage.UserFilters{Email: &email})
		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("error_when_supabase_signup_fails_for_an_invited_user_leaves_it_invited", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignUp", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("supabase unavailable")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		invited := seedUser(t, cfg, "invited_fail@example.com", states.UserStateInvited, constants.RoleTypeReader)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "invited_fail@example.com",
			Password: "S3cr3t-pw",
		})
		require.Error(t, err)

		unchanged, err := cfg.Deps.Storage.User.GetByID(cfg.Ctx, cfg.DB, invited.ID)
		require.NoError(t, err)
		require.Equal(t, states.UserStateInvited, unchanged.EntityState)
		require.Equal(t, invited.RoleID, unchanged.RoleID)
		require.Equal(t, invited.UserRef, unchanged.UserRef)
	})

	t.Run("error_when_user_already_active_does_not_call_supabase", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		authMock := cfg.Deps.SupabaseAuth.(*auth.MockIAuth)
		seedUser(t, cfg, "checked_before_supabase@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:    "checked_before_supabase@example.com",
			Password: "S3cr3t-pw",
		})
		require.Error(t, err)
		authMock.AssertNotCalled(t, "SignUp", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("error_when_password_is_too_short", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "short_pw@example.com",
			Password:     "S3cr3t",
			PracticeName: ptr.To("short_pw_practice"),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least 8 characters")
		authMock.AssertNotCalled(t, "SignUp", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("error_when_password_has_no_uppercase_letter", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "no_upper@example.com",
			Password:     "s3cr3t-pw",
			PracticeName: ptr.To("no_upper_practice"),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "uppercase")
		authMock.AssertNotCalled(t, "SignUp", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("error_when_password_has_no_digit", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, err := bz.User.SignUp(cfg.Ctx, cfg.DB, contracts.SignUpRequest{
			Email:        "no_digit@example.com",
			Password:     "Secretpw",
			PracticeName: ptr.To("no_digit_practice"),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "uppercase")
		authMock.AssertNotCalled(t, "SignUp", mock.Anything, mock.Anything, mock.Anything)
	})
}

func Test_User_Authenticate(t *testing.T) {
	const token = "access-token"

	t.Run("success_returns_the_active_user_matching_the_token_identity", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)

		authMock.On("Authenticate", mock.Anything, token).Return(&auth.AuthenticatedUser{
			ID:    "seeded_supabase_ref",
			Email: "authed_user@example.com",
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		seeded := seedUser(t, cfg, "authed_user@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		user, err := bz.User.Authenticate(cfg.Ctx, cfg.DB, token)
		require.NoError(t, err)
		require.Equal(t, seeded.ID, user.ID)
		require.Equal(t, "seeded_supabase_ref", user.UserRef)
		require.Equal(t, states.UserStateActive, user.EntityState)
	})

	t.Run("error_when_the_token_is_invalid", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("Authenticate", mock.Anything, token).Return(nil, errors.New("invalid token")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, err := bz.User.Authenticate(cfg.Ctx, cfg.DB, token)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid token")
	})

	t.Run("error_when_no_user_matches_the_identity", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("Authenticate", mock.Anything, token).Return(&auth.AuthenticatedUser{
			ID:    "unknown_supabase_ref",
			Email: "ghost@example.com",
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, err := bz.User.Authenticate(cfg.Ctx, cfg.DB, token)
		require.Error(t, err)
	})

	t.Run("error_when_the_matched_user_is_not_active", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("Authenticate", mock.Anything, token).Return(&auth.AuthenticatedUser{
			ID:    "seeded_supabase_ref",
			Email: "disabled_user@example.com",
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		seedUser(t, cfg, "disabled_user@example.com", states.UserStateDisabled, constants.RoleTypeAdmin)

		_, err := bz.User.Authenticate(cfg.Ctx, cfg.DB, token)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not active")
	})
}

func Test_User_SignIn(t *testing.T) {
	t.Run("success_returns_the_active_user_and_session_matching_the_credentials", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)

		authMock.On("SignIn", mock.Anything, "authed_user@example.com", "s3cr3t-pw").Return(&auth.SignInResult{
			AuthenticatedUser: auth.AuthenticatedUser{
				ID:    "seeded_supabase_ref",
				Email: "authed_user@example.com",
			},
			Session: auth.Session{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				TokenType:    "bearer",
				ExpiresIn:    3600,
				ExpiresAt:    1234567890,
			},
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		seeded := seedUser(t, cfg, "authed_user@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		user, session, err := bz.User.SignIn(cfg.Ctx, cfg.DB, contracts.SignInRequest{
			Email:    "authed_user@example.com",
			Password: "s3cr3t-pw",
		})
		require.NoError(t, err)
		require.Equal(t, seeded.ID, user.ID)
		require.Equal(t, "seeded_supabase_ref", user.UserRef)
		require.Equal(t, states.UserStateActive, user.EntityState)
		require.Equal(t, "access-token", session.AccessToken)
		require.Equal(t, "refresh-token", session.RefreshToken)
		require.Equal(t, "bearer", session.TokenType)
		require.Equal(t, 3600, session.ExpiresIn)
		require.Equal(t, int64(1234567890), session.ExpiresAt)
	})

	t.Run("error_when_the_credentials_are_invalid", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignIn", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("invalid credentials")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, _, err := bz.User.SignIn(cfg.Ctx, cfg.DB, contracts.SignInRequest{
			Email:    "authed_user@example.com",
			Password: "wrong-pw",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("error_when_no_user_matches_the_identity", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignIn", mock.Anything, mock.Anything, mock.Anything).Return(&auth.SignInResult{
			AuthenticatedUser: auth.AuthenticatedUser{
				ID:    "unknown_supabase_ref",
				Email: "ghost@example.com",
			},
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, _, err := bz.User.SignIn(cfg.Ctx, cfg.DB, contracts.SignInRequest{
			Email:    "ghost@example.com",
			Password: "s3cr3t-pw",
		})
		require.Error(t, err)
	})

	t.Run("error_when_the_matched_user_is_not_active", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignIn", mock.Anything, mock.Anything, mock.Anything).Return(&auth.SignInResult{
			AuthenticatedUser: auth.AuthenticatedUser{
				ID:    "seeded_supabase_ref",
				Email: "disabled_user@example.com",
			},
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		seedUser(t, cfg, "disabled_user@example.com", states.UserStateDisabled, constants.RoleTypeAdmin)

		_, _, err := bz.User.SignIn(cfg.Ctx, cfg.DB, contracts.SignInRequest{
			Email:    "disabled_user@example.com",
			Password: "s3cr3t-pw",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "not active")
	})
}

func Test_User_Refresh(t *testing.T) {
	t.Run("success_returns_the_active_user_and_rotated_session_matching_the_refresh_token", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)

		authMock.On("RefreshToken", mock.Anything, "old-refresh-token").Return(&auth.SignInResult{
			AuthenticatedUser: auth.AuthenticatedUser{
				ID:    "seeded_supabase_ref",
				Email: "authed_user@example.com",
			},
			Session: auth.Session{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				TokenType:    "bearer",
				ExpiresIn:    3600,
				ExpiresAt:    1234567890,
			},
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		seeded := seedUser(t, cfg, "authed_user@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		user, session, err := bz.User.Refresh(cfg.Ctx, cfg.DB, "old-refresh-token")
		require.NoError(t, err)
		require.Equal(t, seeded.ID, user.ID)
		require.Equal(t, "seeded_supabase_ref", user.UserRef)
		require.Equal(t, states.UserStateActive, user.EntityState)
		require.Equal(t, "new-access-token", session.AccessToken)
		require.Equal(t, "new-refresh-token", session.RefreshToken)
		require.Equal(t, "bearer", session.TokenType)
		require.Equal(t, 3600, session.ExpiresIn)
		require.Equal(t, int64(1234567890), session.ExpiresAt)
	})

	t.Run("error_when_the_refresh_token_is_invalid", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("RefreshToken", mock.Anything, mock.Anything).Return(nil, errors.New("invalid refresh token")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, _, err := bz.User.Refresh(cfg.Ctx, cfg.DB, "bad-refresh-token")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid refresh token")
	})

	t.Run("error_when_no_user_matches_the_identity", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("RefreshToken", mock.Anything, mock.Anything).Return(&auth.SignInResult{
			AuthenticatedUser: auth.AuthenticatedUser{
				ID:    "unknown_supabase_ref",
				Email: "ghost@example.com",
			},
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		_, _, err := bz.User.Refresh(cfg.Ctx, cfg.DB, "some-refresh-token")
		require.Error(t, err)
	})

	t.Run("error_when_the_matched_user_is_not_active", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("RefreshToken", mock.Anything, mock.Anything).Return(&auth.SignInResult{
			AuthenticatedUser: auth.AuthenticatedUser{
				ID:    "seeded_supabase_ref",
				Email: "disabled_user@example.com",
			},
		}, nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		seedUser(t, cfg, "disabled_user@example.com", states.UserStateDisabled, constants.RoleTypeAdmin)

		_, _, err := bz.User.Refresh(cfg.Ctx, cfg.DB, "some-refresh-token")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not active")
	})
}

func Test_User_SignOut(t *testing.T) {
	t.Run("success_revokes_the_session_via_the_access_token", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignOut", mock.Anything, "access-token").Return(nil).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		err := bz.User.SignOut(cfg.Ctx, cfg.DB, "access-token")
		require.NoError(t, err)
	})

	t.Run("error_when_supabase_sign_out_fails", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("SignOut", mock.Anything, mock.Anything).Return(errors.New("supabase unavailable")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)

		err := bz.User.SignOut(cfg.Ctx, cfg.DB, "access-token")
		require.Error(t, err)
		require.Contains(t, err.Error(), "supabase unavailable")
	})
}

func Test_User_Invite(t *testing.T) {
	t.Run("success_creates_an_invited_user_in_the_inviters_practice", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)
		seedRole(t, cfg, constants.RoleTypeReader)
		authMock := cfg.Deps.SupabaseAuth.(*auth.MockIAuth)
		authMock.On("Invite", mock.Anything, "invitee@example.com").
			Return(&auth.InviteResult{ID: "invited_supabase_ref"}, nil).Once()

		user, err := bz.User.Invite(cfg.Ctx, cfg.DB, practice.ID, contracts.InviteUserRequest{Email: "invitee@example.com"})
		require.NoError(t, err)
		require.NotEmpty(t, user.ID)
		require.Equal(t, "invitee@example.com", user.Email)
		require.Equal(t, practice.ID, user.PracticeID)
		require.Equal(t, states.UserStateInvited, user.EntityState)
		require.Equal(t, "invited_supabase_ref", user.UserRef)
	})

	t.Run("error_when_the_user_already_exists_in_the_same_practice", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		existing := seedUser(t, cfg, "dup@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		_, err := bz.User.Invite(cfg.Ctx, cfg.DB, existing.PracticeID, contracts.InviteUserRequest{Email: "dup@example.com"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists in the system")
	})

	t.Run("error_when_the_user_belongs_to_another_practice", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		seedUser(t, cfg, "other@example.com", states.UserStateActive, constants.RoleTypeAdmin)

		_, err := bz.User.Invite(cfg.Ctx, cfg.DB, ksuid.New(), contracts.InviteUserRequest{Email: "other@example.com"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "associated with another practice")
	})

	t.Run("error_when_supabase_invite_fails_leaves_no_user_created", func(t *testing.T) {
		authMock := auth.NewMockIAuth(t)
		authMock.On("Invite", mock.Anything, mock.Anything).Return(nil, errors.New("supabase unavailable")).Once()

		cfg, bz := newUserBizWithAuthMock(t, authMock)
		practice := fixtures.NewPractice(t, cfg, bz)
		seedRole(t, cfg, constants.RoleTypeReader)

		_, err := bz.User.Invite(cfg.Ctx, cfg.DB, practice.ID, contracts.InviteUserRequest{Email: "invitee@example.com"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "supabase unavailable")

		email := "invitee@example.com"
		users, err := cfg.Deps.Storage.User.Get(cfg.Ctx, cfg.DB, storage.UserFilters{Email: &email})
		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("error_when_no_active_reader_role_exists", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)
		clearRoleType(t, cfg, constants.RoleTypeReader)

		_, err := bz.User.Invite(cfg.Ctx, cfg.DB, practice.ID, contracts.InviteUserRequest{Email: "invitee@example.com"})
		require.Error(t, err)
	})
}

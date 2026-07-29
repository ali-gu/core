package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_InviteUser(t *testing.T) {
	t.Run("201_invites_a_user_into_the_inviters_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		env.seedRole(constants.RoleTypeReader)
		env.authMock.On("Invite", mock.Anything, "invitee@example.com").
			Return(&auth.InviteResult{ID: "invited_supabase_ref"}, nil)

		w := env.do(http.MethodPost, "/v1/users/invite", token, contracts.InviteUserRequest{Email: "invitee@example.com"})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.InviteUserResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "invitee@example.com", resp.Email)
		require.Equal(t, practice.ID, resp.PracticeID)
		require.Equal(t, states.UserStateInvited, resp.EntityState)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users/invite", "", contracts.InviteUserRequest{Email: "invitee@example.com"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400_when_the_email_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/users/invite", token, contracts.InviteUserRequest{})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_the_email_is_malformed", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/users/invite", token, contracts.InviteUserRequest{Email: "not-an-email"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

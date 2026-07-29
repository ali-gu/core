package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/stretchr/testify/require"
)

func Test_SignUp(t *testing.T) {
	t.Run("201_creates_a_new_user_and_practice", func(t *testing.T) {
		env := newE2E(t)

		env.seedRole(constants.RoleTypeAdmin)

		w := env.do(http.MethodPost, "/v1/users", "", contracts.SignUpRequest{
			Email:        "owner@example.com",
			Password:     "S3cr3t-pw",
			PracticeName: ptr.To("Bright Smiles"),
		})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.SignUpResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "owner@example.com", resp.Email)
		require.Equal(t, states.UserStateActive, resp.EntityState)
		require.NotEmpty(t, resp.PracticeID)
	})

	t.Run("400_when_the_email_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users", "", contracts.SignUpRequest{Password: "s3cr3t-pw"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_the_email_is_malformed", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users", "", contracts.SignUpRequest{
			Email:    "not-an-email",
			Password: "s3cr3t-pw",
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_the_password_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/users", "", contracts.SignUpRequest{Email: "owner@example.com"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

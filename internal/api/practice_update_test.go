package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_UpdatePractice(t *testing.T) {
	t.Run("200_updates_the_authenticated_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		w := env.do(http.MethodPatch, "/v1/practices/"+practice.ID.String(), token, contracts.UpdatePracticeRequest{
			Email:   ptr.To("ops@example.com"),
			ZipCode: ptr.To("95014"),
			Website: ptr.To("www.example.com"),
		})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.UpdatePracticeResponse
		decodeBody(t, w, &resp)
		require.Equal(t, practice.ID, resp.ID)
		require.NotNil(t, resp.Email)
		require.Equal(t, "ops@example.com", *resp.Email)
		require.NotNil(t, resp.ZipCode)
		require.Equal(t, "95014", *resp.ZipCode)
		require.NotNil(t, resp.Website)
		require.Equal(t, "www.example.com", *resp.Website)
	})

	t.Run("200_updates_only_the_given_field", func(t *testing.T) {
		env := newE2E(t)
		practice := fixtures.NewPractice(t, env.cfg, env.bz, fixtures.WithPracticeNoContactInfo())
		token := env.authFor(practice.ID)

		w := env.do(http.MethodPatch, "/v1/practices/"+practice.ID.String(), token, contracts.UpdatePracticeRequest{
			Email: ptr.To("ops@example.com"),
		})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.UpdatePracticeResponse
		decodeBody(t, w, &resp)
		require.NotNil(t, resp.Email)
		require.Equal(t, "ops@example.com", *resp.Email)
		require.Nil(t, resp.ZipCode)
		require.Nil(t, resp.Website)
	})

	t.Run("403_when_updating_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherPractice := fixtures.NewPractice(t, env.cfg, env.bz)

		w := env.do(http.MethodPatch, "/v1/practices/"+otherPractice.ID.String(), token, contracts.UpdatePracticeRequest{
			Email: ptr.To("ops@example.com"),
		})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPatch, fmt.Sprintf("/v1/practices/%s", ksuid.New()), "", contracts.UpdatePracticeRequest{})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_DisablePhoneNumber(t *testing.T) {
	t.Run("200_disables_an_active_phone_number", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		pn := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/phone-numbers/"+pn.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.DisablePhoneNumberResponse
		decodeBody(t, w, &resp)
		require.Equal(t, pn.ID, resp.ID)
		require.Equal(t, "DISABLED", resp.EntityState)
		require.NotNil(t, resp.DisabledAt)
	})

	t.Run("error_disabling_a_reserved_phone_number", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		pn, err := env.bz.PhoneNumber.Reserve(env.cfg.Ctx, env.cfg.DB, practice.ID,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550199"})
		require.NoError(t, err)

		w := env.do(http.MethodPatch, "/v1/phone-numbers/"+pn.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error_disabling_an_already_disabled_phone_number", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		location := fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID))
		pn := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/phone-numbers/"+pn.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		w = env.do(http.MethodPatch, "/v1/phone-numbers/"+pn.ID.String()+"/disable", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_phone_number_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		owningPractice, _ := env.authedPractice()
		_, otherToken := env.authedPractice()

		location := fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(owningPractice.ID))
		pn := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(location.ID))

		w := env.do(http.MethodPatch, "/v1/phone-numbers/"+pn.ID.String()+"/disable", otherToken, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPatch, "/v1/phone-numbers/"+ksuid.New().String()+"/disable", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

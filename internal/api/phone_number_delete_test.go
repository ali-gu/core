package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_DeletePhoneNumber(t *testing.T) {
	t.Run("204_when_the_phone_number_belongs_to_the_users_practice_and_is_not_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		pn, err := env.bz.PhoneNumber.Reserve(env.cfg.Ctx, env.cfg.DB, practice.ID,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550111"})
		require.NoError(t, err)

		w := env.do(http.MethodDelete, "/v1/phone-numbers/"+pn.ID.String(), token, nil)
		require.Equal(t, http.StatusNoContent, w.Code)

		numbers, err := env.bz.PhoneNumber.Get(env.cfg.Ctx, env.cfg.DB, practice.ID)
		require.NoError(t, err)
		for _, n := range numbers {
			require.NotEqual(t, pn.ID, n.ID)
		}
	})

	t.Run("403_when_the_phone_number_is_active", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()
		pn := fixtures.NewPhoneNumber(t, env.cfg, env.bz, fixtures.WithPhoneNumberLocationID(
			fixtures.NewLocation(t, env.cfg, env.bz, fixtures.WithLocationPracticeID(practice.ID)).ID))

		w := env.do(http.MethodDelete, "/v1/phone-numbers/"+pn.ID.String(), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("403_when_the_phone_number_belongs_to_another_practice", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()
		otherPractice, _ := env.authedPractice()

		pn, err := env.bz.PhoneNumber.Reserve(env.cfg.Ctx, env.cfg.DB, otherPractice.ID,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550112"})
		require.NoError(t, err)

		w := env.do(http.MethodDelete, "/v1/phone-numbers/"+pn.ID.String(), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodDelete, fmt.Sprintf("/v1/phone-numbers/%s", ksuid.New()), "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("403_when_the_phone_number_does_not_exist", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodDelete, fmt.Sprintf("/v1/phone-numbers/%s", ksuid.New()), token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

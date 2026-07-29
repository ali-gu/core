package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/stretchr/testify/require"
)

func Test_ReservePhoneNumber(t *testing.T) {
	t.Run("201_reserves_a_number_in_the_authenticated_users_practice", func(t *testing.T) {
		env := newE2E(t)
		practice, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/phone-numbers/reserve", token,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550101"})
		require.Equal(t, http.StatusCreated, w.Code)

		var resp contracts.ReservePhoneNumberResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "+15555550101", resp.PhoneNumber)

		numbers, err := env.cfg.Deps.Storage.PhoneNumber.Get(env.cfg.Ctx, env.cfg.DB, storage.PhoneNumberFilters{PracticeID: &practice.ID})
		require.NoError(t, err)
		require.Len(t, numbers, 1)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/phone-numbers/reserve", "",
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550101"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400_when_the_phone_number_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/phone-numbers/reserve", token, contracts.ReservePhoneNumberRequest{})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("403_when_the_practice_already_has_a_reserved_phone_number", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/phone-numbers/reserve", token,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550101"})
		require.Equal(t, http.StatusCreated, w.Code)

		w = env.do(http.MethodPost, "/v1/phone-numbers/reserve", token,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550102"})
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func Test_ActivatePhoneNumber(t *testing.T) {
	t.Run("200_activates_a_reserved_number", func(t *testing.T) {
		env := newE2E(t)
		_, adminToken := env.authedSuperAdminPractice()
		_, userToken := env.authedPractice()

		reserveW := env.do(http.MethodPost, "/v1/phone-numbers/reserve", userToken,
			contracts.ReservePhoneNumberRequest{PhoneNumber: "+15555550102"})
		require.Equal(t, http.StatusCreated, reserveW.Code)

		var reserveResp contracts.ReservePhoneNumberResponse
		decodeBody(t, reserveW, &reserveResp)
		require.Equal(t, "RESERVED", reserveResp.EntityState)

		w := env.do(http.MethodPost, "/v1/admin/phone-numbers/activate", adminToken,
			contracts.ActivatePhoneNumberRequest{PhoneNumberID: reserveResp.ID, PhoneNumberRef: "telnyx_number_ref"})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.ActivatePhoneNumberResponse
		decodeBody(t, w, &resp)
		require.Equal(t, reserveResp.ID, resp.ID)
		require.Equal(t, "ACTIVE", resp.EntityState)
		require.NotNil(t, resp.PhoneNumberIDRef)
		require.Equal(t, "telnyx_number_ref", *resp.PhoneNumberIDRef)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodPost, "/v1/admin/phone-numbers/activate", "",
			contracts.ActivatePhoneNumberRequest{PhoneNumberID: ksuid.New(), PhoneNumberRef: "telnyx_number_ref"})
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("403_when_the_user_is_not_a_super_admin", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodPost, "/v1/admin/phone-numbers/activate", token,
			contracts.ActivatePhoneNumberRequest{PhoneNumberID: ksuid.New(), PhoneNumberRef: "telnyx_number_ref"})
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("400_when_the_phone_number_id_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/admin/phone-numbers/activate", token,
			contracts.ActivatePhoneNumberRequest{PhoneNumberRef: "telnyx_number_ref"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400_when_the_phone_number_ref_is_missing", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		w := env.do(http.MethodPost, "/v1/admin/phone-numbers/activate", token,
			contracts.ActivatePhoneNumberRequest{PhoneNumberID: ksuid.New()})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

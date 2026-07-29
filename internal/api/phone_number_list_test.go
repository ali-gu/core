package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_ListPurchasedPhoneNumbers(t *testing.T) {
	t.Run("200_returns_purchased_numbers_from_the_provider", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedSuperAdminPractice()

		phoneMock := env.cfg.Deps.TelnyxPhoneNumberManager.(*phonenumber.MockIPhoneNumberManager)
		phoneMock.On("ListPurchased", mock.Anything).Return([]phonenumber.PurchasedPhoneNumber{
			{ID: "pn_1", PhoneNumber: "+15555550100", Status: "active"},
		}, nil)

		w := env.do(http.MethodGet, "/v1/admin/phone-numbers/purchased", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.ListPurchasedPhoneNumbersResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.Equal(t, "pn_1", resp.Data[0].ID)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/admin/phone-numbers/purchased", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("403_when_the_user_is_not_a_super_admin", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		w := env.do(http.MethodGet, "/v1/admin/phone-numbers/purchased", token, nil)
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}

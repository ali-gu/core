package api_test

import (
	"net/http"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_ListAvailablePhoneNumbers(t *testing.T) {
	t.Run("200_returns_available_numbers_from_the_provider", func(t *testing.T) {
		env := newE2E(t)
		_, token := env.authedPractice()

		phoneMock := env.cfg.Deps.TelnyxPhoneNumberManager.(*phonenumber.MockIPhoneNumberManager)
		phoneMock.On("ListAvailable", mock.Anything, mock.Anything).Return([]phonenumber.AvailablePhoneNumber{
			{PhoneNumber: "+14155550100", Reservable: true},
		}, nil)

		w := env.do(http.MethodGet, "/v1/phone-numbers/available?country_code=US&area_code=415", token, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.ListAvailablePhoneNumbersResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.Equal(t, "+14155550100", resp.Data[0].PhoneNumber)
	})

	t.Run("401_without_a_token", func(t *testing.T) {
		env := newE2E(t)

		w := env.do(http.MethodGet, "/v1/phone-numbers/available", "", nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

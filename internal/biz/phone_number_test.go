package biz_test

import (
	"errors"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_PhoneNumber_Get(t *testing.T) {
	t.Run("success_empty", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		phoneNumbers, err := bz.PhoneNumber.Get(cfg.Ctx, cfg.DB, ksuid.New())
		require.NoError(t, err)
		require.Empty(t, phoneNumbers)
	})

	t.Run("success_orders_active_before_reserved", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, location.PracticeID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.NoError(t, err)

		active := fixtures.NewPhoneNumber(t, cfg, bz,
			fixtures.WithPhoneNumberLocationID(location.ID), fixtures.WithPhoneNumberNumber("+15555550100"))

		phoneNumbers, err := bz.PhoneNumber.Get(cfg.Ctx, cfg.DB, location.PracticeID)
		require.NoError(t, err)
		require.Len(t, phoneNumbers, 2)
		require.Equal(t, states.PhoneNumberStateActive, phoneNumbers[0].EntityState)
		require.Equal(t, active.ID, phoneNumbers[0].ID)
		require.Equal(t, states.PhoneNumberStateReserved, phoneNumbers[1].EntityState)
		require.Equal(t, reserved.ID, phoneNumbers[1].ID)
	})

	t.Run("success_returns_only_the_given_practices_numbers", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		fixtures.NewPhoneNumber(t, cfg, bz)

		phoneNumbers, err := bz.PhoneNumber.Get(cfg.Ctx, cfg.DB, location.PracticeID)
		require.NoError(t, err)
		require.Len(t, phoneNumbers, 1)
		require.Equal(t, location.PracticeID, phoneNumbers[0].PracticeID)
	})
}

func Test_PhoneNumber_ListPurchased(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		phoneMock := cfg.Deps.TelnyxPhoneNumberManager.(*phonenumber.MockIPhoneNumberManager)
		phoneMock.On("ListPurchased", mock.Anything).Return([]phonenumber.PurchasedPhoneNumber{
			{ID: "pn_1", PhoneNumber: "+15555550100", Status: "active"},
		}, nil)

		result, err := bz.PhoneNumber.ListPurchased(cfg.Ctx, cfg.DB)
		require.NoError(t, err)
		require.Equal(t, []phonenumber.PurchasedPhoneNumber{
			{ID: "pn_1", PhoneNumber: "+15555550100", Status: "active"},
		}, result)
	})

	t.Run("error_propagated_from_telnyx", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		phoneMock := cfg.Deps.TelnyxPhoneNumberManager.(*phonenumber.MockIPhoneNumberManager)
		phoneMock.On("ListPurchased", mock.Anything).Return(nil, errors.New("telnyx unavailable"))

		_, err := bz.PhoneNumber.ListPurchased(cfg.Ctx, cfg.DB)
		require.Error(t, err)
		require.Contains(t, err.Error(), "telnyx unavailable")
	})
}

func Test_PhoneNumber_ListAvailable(t *testing.T) {
	t.Run("success_passes_params_through_to_telnyx", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		phoneMock := cfg.Deps.TelnyxPhoneNumberManager.(*phonenumber.MockIPhoneNumberManager)
		phoneMock.On("ListAvailable", mock.Anything, phonenumber.ListAvailablePhoneNumbersParams{
			CountryCode: "US",
			AreaCode:    "415",
			Contains:    "555",
		}).Return([]phonenumber.AvailablePhoneNumber{
			{PhoneNumber: "+14155550100", Reservable: true},
		}, nil)

		result, err := bz.PhoneNumber.ListAvailable(cfg.Ctx, cfg.DB, contracts.ListAvailablePhoneNumbersRequest{
			CountryCode: "US",
			AreaCode:    "415",
			Contains:    "555",
		})
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "+14155550100", result[0].PhoneNumber)
	})

	t.Run("error_propagated_from_telnyx", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		phoneMock := cfg.Deps.TelnyxPhoneNumberManager.(*phonenumber.MockIPhoneNumberManager)
		phoneMock.On("ListAvailable", mock.Anything, mock.Anything).Return(nil, errors.New("telnyx unavailable"))

		_, err := bz.PhoneNumber.ListAvailable(cfg.Ctx, cfg.DB, contracts.ListAvailablePhoneNumbersRequest{})
		require.Error(t, err)
	})
}

func Test_PhoneNumber_Reserve(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practice.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.NoError(t, err)
		require.Equal(t, "+15555550101", reserved.PhoneNumber)
		require.Equal(t, practice.ID, reserved.PracticeID)
		require.Equal(t, states.PhoneNumberStateReserved, reserved.EntityState)
		require.NotNil(t, reserved.PhoneNumberReservationRef)
		require.Nil(t, reserved.PhoneNumberIDRef)
	})

	t.Run("error_when_practice_already_has_a_reserved_phone_number", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		_, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practice.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.NoError(t, err)

		_, err = bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practice.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550102",
		})
		require.Error(t, err)
		var re *rerror.Error
		require.ErrorAs(t, err, &re)
		require.Equal(t, rerror.Forbidden, re.Kind())
	})

	t.Run("success_allows_reserving_when_practice_already_has_an_active_phone_number", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)
		fixtures.NewPhoneNumber(t, cfg, bz, fixtures.WithPhoneNumberLocationID(location.ID))

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, location.PracticeID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.NoError(t, err)
		require.Equal(t, "+15555550101", reserved.PhoneNumber)
	})

	t.Run("success_allows_reserving_after_the_previous_reservation_was_deleted", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practice := fixtures.NewPractice(t, cfg, bz)

		first, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practice.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.NoError(t, err)

		require.NoError(t, bz.PhoneNumber.Delete(cfg.Ctx, cfg.DB, practice.ID, first.ID))

		reserved, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practice.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550102",
		})
		require.NoError(t, err)
		require.Equal(t, "+15555550102", reserved.PhoneNumber)
	})

	t.Run("success_allows_different_practices_to_each_reserve_a_number", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		practiceA := fixtures.NewPractice(t, cfg, bz)
		practiceB := fixtures.NewPractice(t, cfg, bz)

		_, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practiceA.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.NoError(t, err)

		_, err = bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, practiceB.ID, contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550102",
		})
		require.NoError(t, err)
	})

	t.Run("error_when_practice_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.PhoneNumber.Reserve(cfg.Ctx, cfg.DB, ksuid.New(), contracts.ReservePhoneNumberRequest{
			PhoneNumber: "+15555550101",
		})
		require.Error(t, err)
	})
}

package biz_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Appointment_Find(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		ehrRecord := fixtures.NewEHR(t, cfg, bz,
			fixtures.WithEHRLocationID(location.ID),
			fixtures.WithEHRLocationRef("nexhealth_location_id"),
		)

		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		startDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		expected := []ehr.Appointment{
			{Time: startDate.Add(time.Hour), ProviderID: "provider_1", ProviderName: "Dr. Smith"},
		}

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("FindAppointment", mock.Anything, ehr.AvailableAppointmentsParams{
			Subdomain:  ehrRecord.Subdomain,
			LocationID: *ehrRecord.LocationRef,
			StartDate:  startDate,
			Days:       constants.DefaultAppointmentSearchDays,
		}).Return(expected, nil)

		result, err := bz.Appointment.Find(cfg.Ctx, cfg.DB, contracts.FindAppointmentRequest{
			AssistantID: "assistant_1",
			StartDate:   startDate,
		})
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("error_propagated_from_ehr", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("FindAppointment", mock.Anything, mock.Anything).Return(nil, errors.New("nexhealth unavailable"))

		_, err := bz.Appointment.Find(cfg.Ctx, cfg.DB, contracts.FindAppointmentRequest{
			AssistantID: "assistant_1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "nexhealth unavailable")
	})

	t.Run("error_when_assistant_id_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Appointment.Find(cfg.Ctx, cfg.DB, contracts.FindAppointmentRequest{
			AssistantID: "does_not_exist",
		})
		require.Error(t, err)
	})

	t.Run("error_when_agent_has_no_location_assigned", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewLocation(t, cfg, bz)

		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentPracticeID(location.PracticeID),
			fixtures.WithAgentRef("assistant_1"),
		)

		_, err := bz.Appointment.Find(cfg.Ctx, cfg.DB, contracts.FindAppointmentRequest{
			AssistantID: "assistant_1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "no location assigned")
	})

	t.Run("error_when_location_has_no_ehr", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		agentRecord := fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentPracticeID(location.PracticeID),
			fixtures.WithAgentRef("assistant_1"),
		)

		agentRecord.AssignLocation(location.ID)
		require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, agentRecord))

		_, err := bz.Appointment.Find(cfg.Ctx, cfg.DB, contracts.FindAppointmentRequest{
			AssistantID: "assistant_1",
		})
		require.Error(t, err)
	})

	t.Run("error_when_ehr_is_missing_nexhealth_configuration", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		agentRecord := fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentPracticeID(location.PracticeID),
			fixtures.WithAgentRef("assistant_1"),
		)

		agentRecord.AssignLocation(location.ID)
		require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, agentRecord))

		require.NoError(t, cfg.Deps.Storage.EHR.Create(cfg.Ctx, cfg.DB, storage.EHRS{
			ID:         ksuid.New(),
			Type:       constants.EHRNexHealth,
			LocationID: location.ID,
			CreatedAt:  time.Now(),
		}))

		_, err := bz.Appointment.Find(cfg.Ctx, cfg.DB, contracts.FindAppointmentRequest{
			AssistantID: "assistant_1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing nexhealth configuration")
	})
}

func Test_Appointment_Book(t *testing.T) {
	newRequest := func(assistantID string) contracts.BookAppointmentRequest {
		return contracts.BookAppointmentRequest{
			AssistantID: assistantID,
			DoctorID:    "104",
			Time:        time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC),
			Reason:      "cleaning",
			Patient: contracts.PatientContact{
				FirstName:   "John",
				LastName:    "Smith",
				Phone:       "+15555559999",
				DateOfBirth: "1990-05-03",
			},
		}
	}

	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		ehrRecord := fixtures.NewEHR(t, cfg, bz,
			fixtures.WithEHRLocationID(location.ID),
			fixtures.WithEHRLocationRef("nexhealth_location_id"),
		)
		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		expected := &ehr.BookedAppointment{
			ID:         "1822",
			Time:       time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC),
			ProviderID: "104",
			Confirmed:  true,
		}

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("BookAppointment", mock.Anything, ehr.BookAppointmentParams{
			Subdomain:  ehrRecord.Subdomain,
			LocationID: *ehrRecord.LocationRef,
			ProviderID: "104",
			Time:       time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC),
			Note:       "cleaning",
			Patient: ehr.PatientDetails{
				FirstName:   "John",
				LastName:    "Smith",
				Phone:       "+15555559999",
				DateOfBirth: "1990-05-03",
			},
		}).Return(expected, nil)

		result, err := bz.Appointment.Book(cfg.Ctx, cfg.DB, newRequest("assistant_1"))
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("error_propagated_from_ehr", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("BookAppointment", mock.Anything, mock.Anything).Return(nil, errors.New("nexhealth unavailable"))

		_, err := bz.Appointment.Book(cfg.Ctx, cfg.DB, newRequest("assistant_1"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "nexhealth unavailable")
	})

	t.Run("error_when_assistant_id_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Appointment.Book(cfg.Ctx, cfg.DB, newRequest("does_not_exist"))
		require.Error(t, err)
	})
}

func Test_Appointment_Cancel(t *testing.T) {
	newRequest := func(assistantID string) contracts.CancelAppointmentRequest {
		return contracts.CancelAppointmentRequest{
			AssistantID: assistantID,
			Patient: contracts.PatientIdentity{
				Name:        "John Smith",
				Phone:       "+15555559999",
				DateOfBirth: "1990-05-03",
			},
		}
	}

	t.Run("success", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		ehrRecord := fixtures.NewEHR(t, cfg, bz,
			fixtures.WithEHRLocationID(location.ID),
			fixtures.WithEHRLocationRef("nexhealth_location_id"),
		)
		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		expected := &ehr.CancelledAppointment{
			ID:         "1822",
			Time:       time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC),
			ProviderID: "104",
		}

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CancelAppointment", mock.Anything, ehr.CancelAppointmentParams{
			Subdomain:  ehrRecord.Subdomain,
			LocationID: *ehrRecord.LocationRef,
			Patient: ehr.PatientLookup{
				Name:        "John Smith",
				Phone:       "+15555559999",
				DateOfBirth: "1990-05-03",
			},
		}).Return(expected, nil)

		result, err := bz.Appointment.Cancel(cfg.Ctx, cfg.DB, newRequest("assistant_1"))
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("error_propagated_from_ehr", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)
		location := fixtures.NewPendingLocation(t, cfg, bz)
		fixtures.NewEHR(t, cfg, bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, cfg, bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CancelAppointment", mock.Anything, mock.Anything).Return(nil, errors.New("nexhealth unavailable"))

		_, err := bz.Appointment.Cancel(cfg.Ctx, cfg.DB, newRequest("assistant_1"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "nexhealth unavailable")
	})

	t.Run("error_when_assistant_id_does_not_exist", func(t *testing.T) {
		cfg, bz := testutils.BasicSetup(t)

		_, err := bz.Appointment.Cancel(cfg.Ctx, cfg.DB, newRequest("does_not_exist"))
		require.Error(t, err)
	})
}

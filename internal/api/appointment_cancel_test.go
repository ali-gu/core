package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_CancelAppointment(t *testing.T) {
	t.Run("200_cancels_and_returns_the_appointment", func(t *testing.T) {
		env := newE2E(t)

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz)
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		cancelledTime := time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC)
		ehrMock := env.cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("CancelAppointment", mock.Anything, mock.Anything).Return(&ehr.CancelledAppointment{
			ID:         "1822",
			Time:       cancelledTime,
			ProviderID: "104",
		}, nil)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/appointments/cancel", contracts.CancelAppointmentRequest{
			AssistantID: "assistant_1",
			Patient: contracts.PatientIdentity{
				Name:        "John Smith",
				Phone:       "+15555559999",
				DateOfBirth: "1990-05-03",
			},
		})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.CancelAppointmentResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "1822", resp.Data.AppointmentID)
		require.Equal(t, "104", resp.Data.DoctorID)
	})

	t.Run("400_when_patient_has_no_phone_or_date_of_birth", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/appointments/cancel", contracts.CancelAppointmentRequest{
			AssistantID: "assistant_1",
			Patient: contracts.PatientIdentity{
				Name: "John Smith",
			},
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

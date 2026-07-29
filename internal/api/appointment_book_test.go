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

func Test_BookAppointment(t *testing.T) {
	t.Run("200_books_and_returns_confirmation", func(t *testing.T) {
		env := newE2E(t)

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz)
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		bookedTime := time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC)
		ehrMock := env.cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("BookAppointment", mock.Anything, mock.Anything).Return(&ehr.BookedAppointment{
			ID:         "1822",
			Time:       bookedTime,
			ProviderID: "104",
			Confirmed:  true,
		}, nil)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/appointments/book", contracts.BookAppointmentRequest{
			AssistantID: "assistant_1",
			DoctorID:    "104",
			Time:        bookedTime,
			Reason:      "cleaning",
			Patient: contracts.PatientContact{
				FirstName:   "John",
				LastName:    "Smith",
				Phone:       "+15555559999",
				DateOfBirth: "1990-05-03",
			},
		})
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.BookAppointmentResponse
		decodeBody(t, w, &resp)
		require.Equal(t, "1822", resp.Data.AppointmentID)
		require.Equal(t, "104", resp.Data.ProviderID)
		require.True(t, resp.Data.Confirmed)
	})

	t.Run("400_when_patient_details_are_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodPost, "/v1/webhooks/appointments/book", contracts.BookAppointmentRequest{
			AssistantID: "assistant_1",
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

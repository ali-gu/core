package api_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_FindAppointment(t *testing.T) {
	t.Run("200_returns_appointments_and_needs_no_auth", func(t *testing.T) {
		env := newE2E(t)

		location := fixtures.NewPendingLocation(t, env.cfg, env.bz)
		fixtures.NewEHR(t, env.cfg, env.bz, fixtures.WithEHRLocationID(location.ID))
		fixtures.NewAgent(t, env.cfg, env.bz,
			fixtures.WithAgentLocationID(location.ID),
			fixtures.WithAgentRef("assistant_1"),
		)

		ehrMock := env.cfg.Deps.EHR.(*ehr.MockIEHR)
		ehrMock.On("FindAppointment", mock.Anything, mock.Anything).Return([]ehr.Appointment{
			{Time: time.Now(), ProviderID: "prov_1", ProviderName: "Dr. Smith"},
		}, nil)

		q := url.Values{}
		q.Set("start_date", "2026-07-10")
		q.Set("assistant_id", "assistant_1")

		w := env.doWebhook(http.MethodGet, "/v1/webhooks/appointments/find?"+q.Encode(), nil)
		require.Equal(t, http.StatusOK, w.Code)

		var resp contracts.FindAppointmentsResponse
		decodeBody(t, w, &resp)
		require.Len(t, resp.Data, 1)
		require.Equal(t, "Dr. Smith", resp.Data[0].DoctorsName)
		require.Equal(t, "prov_1", resp.Data[0].DoctorID)
	})

	t.Run("400_when_the_assistant_id_query_param_is_missing", func(t *testing.T) {
		env := newE2E(t)

		w := env.doWebhook(http.MethodGet, "/v1/webhooks/appointments/find", nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

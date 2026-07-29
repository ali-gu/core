package nexhealth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr/nexhealth"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/stretchr/testify/require"
)

func newTestNexHealth(t *testing.T, handler http.HandlerFunc) (func() *nexhealth.NexHealth, storage.DB, *int32) {
	t.Helper()

	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	cfg, err := internal.NewConfig(constants.LandscapeTest)
	require.NoError(t, err)

	ctx := context.Background()
	dbMux, err := storage.NewDBMux(ctx, *cfg)
	require.NoError(t, err)
	t.Cleanup(dbMux.Close)

	db, err := dbMux.BeginWrite()
	require.NoError(t, err)

	authStorage := &storage.NexHealthAuthStorage{}
	require.NoError(t, authStorage.Update(ctx, db, "", time.Unix(0, 0)))
	t.Cleanup(func() {
		require.NoError(t, authStorage.Update(context.Background(), db, "", time.Unix(0, 0)))
	})

	newClient := func() *nexhealth.NexHealth {
		return nexhealth.New(internal.NexHealthConfig{
			BaseURL: server.URL,
			APIKey:  "test-api-key",
		}, dbMux, authStorage)
	}

	return newClient, db, &callCount
}

func jsonToken(token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": true,
			"data": map[string]any{"token": token},
		})
	}
}

func Test_NexHealth_Authenticate_SharesTokenAcrossInstances(t *testing.T) {
	newClient, _, callCount := newTestNexHealth(t, jsonToken("shared-token"))

	firstInstance := newClient()
	require.NoError(t, firstInstance.Authenticate(context.Background()))
	require.EqualValues(t, 1, atomic.LoadInt32(callCount))

	secondInstance := newClient()
	require.NoError(t, secondInstance.Authenticate(context.Background()))
	require.EqualValues(t, 1, atomic.LoadInt32(callCount))
}

func Test_NexHealth_Authenticate_RefreshesExpiredToken(t *testing.T) {
	newClient, db, callCount := newTestNexHealth(t, jsonToken("new-token"))
	ctx := context.Background()

	authStorage := &storage.NexHealthAuthStorage{}
	require.NoError(t, authStorage.Update(ctx, db, "stale-token", time.Now().Add(-time.Hour)))

	client := newClient()
	require.NoError(t, client.Authenticate(ctx))
	require.EqualValues(t, 1, atomic.LoadInt32(callCount))
}

func Test_NexHealth_Authenticate_WaitsForConcurrentRefreshInsteadOfDoubleAuthenticating(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": true,
			"data": map[string]any{"token": "token-" + strconv.FormatInt(time.Now().UnixNano(), 10)},
		})
	}
	newClient, _, callCount := newTestNexHealth(t, handler)

	clientA := newClient()
	clientB := newClient()

	start := make(chan struct{})
	errs := make([]error, 2)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		errs[0] = clientA.Authenticate(context.Background())
	}()
	go func() {
		defer wg.Done()
		<-start
		errs[1] = clientB.Authenticate(context.Background())
	}()
	close(start)
	wg.Wait()

	require.NoError(t, errs[0])
	require.NoError(t, errs[1])
	require.EqualValues(t, 1, atomic.LoadInt32(callCount))
}

func Test_NexHealth_FindAppointment_ReturnsProviderNames(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/authenticates":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": true,
				"data": map[string]any{"token": "test-token"},
			})
		case "/providers":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": true,
				"data": []map[string]any{
					{"id": 101, "name": "John Smith", "display_name": "Dr. Smith", "inactive": false},
					{"id": 102, "name": "Jane Doe", "display_name": "", "inactive": false},
					{"id": 103, "name": "Retired Doc", "display_name": "", "inactive": true},
				},
			})
		case "/appointment_slots":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":  true,
				"count": 2,
				"data": []map[string]any{
					{
						"lid": 1, "pid": 101, "operatory_id": 1,
						"slots": []map[string]any{
							{"time": "2026-08-01T15:00:00.000-00:00", "operatory_id": 1, "provider_id": 101},
						},
					},
					{
						"lid": 1, "pid": 102, "operatory_id": 1,
						"slots": []map[string]any{
							{"time": "2026-08-01T16:00:00.000-00:00", "operatory_id": 1, "provider_id": 102},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}

	newClient, _, _ := newTestNexHealth(t, handler)
	client := newClient()

	appointments, err := client.FindAppointment(context.Background(), ehr.AvailableAppointmentsParams{
		Subdomain:  "test-subdomain",
		LocationID: "1",
		StartDate:  time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Days:       7,
	})
	require.NoError(t, err)
	require.Len(t, appointments, 2)

	byProvider := make(map[string]ehr.Appointment, len(appointments))
	for _, a := range appointments {
		byProvider[a.ProviderID] = a
	}

	require.Equal(t, "Dr. Smith", byProvider["101"].ProviderName)
	require.Equal(t, "Jane Doe", byProvider["102"].ProviderName)
}

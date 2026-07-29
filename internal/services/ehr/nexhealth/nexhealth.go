package nexhealth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"

	"github.com/go-resty/resty/v2"
)

const (
	acceptHeader          = "application/vnd.Nexhealth+json;version=2"
	tokenSafetyMargin     = 60 * time.Second
	fallbackTokenTTL      = 50 * time.Minute
	requestTimeout        = 30 * time.Second
	defaultSearchDays     = 7
	onboardingsAPIVersion = "v20240412"

	appointmentsAPIVersion = "v3.0.0"

	appointmentsPageSize = 20

	cancelSearchWindow = 365 * 24 * time.Hour
)

type authResponse struct {
	Code bool `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type appointmentSlotsResponse struct {
	Code  bool `json:"code"`
	Count int  `json:"count"`
	Data  []struct {
		LID         int `json:"lid"`
		PID         int `json:"pid"`
		OperatoryID int `json:"operatory_id"`
		Slots       []struct {
			Time         string `json:"time"`
			OperatoryID  int    `json:"operatory_id"`
			ProviderID   int    `json:"provider_id"`
			ProviderName string `json:"-"`
		} `json:"slots"`
	} `json:"data"`
}

func (r *appointmentSlotsResponse) setProviderNames(providerNames map[string]string) {
	for i, group := range r.Data {
		for j, slot := range group.Slots {
			providerID := slot.ProviderID
			if providerID == 0 {
				providerID = group.PID
			}
			r.Data[i].Slots[j].ProviderName = providerNames[strconv.Itoa(providerID)]
		}
	}
}

func (r appointmentSlotsResponse) toAppointments() ([]ehr.Appointment, error) {
	var appointments []ehr.Appointment
	for _, group := range r.Data {
		for _, slot := range group.Slots {
			at, err := time.Parse(time.RFC3339, slot.Time)
			if err != nil {
				return nil, rerror.New(fmt.Errorf("nexhealth: find appointment: parse time %q: %w", slot.Time, err))
			}

			providerID := slot.ProviderID
			if providerID == 0 {
				providerID = group.PID
			}

			appointments = append(appointments, ehr.Appointment{
				Time:         at,
				ProviderID:   strconv.Itoa(providerID),
				ProviderName: slot.ProviderName,
			})
		}
	}
	return appointments, nil
}

type providersResponse struct {
	Code bool `json:"code"`
	Data []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Inactive    bool   `json:"inactive"`
	} `json:"data"`
}

type nexProvider struct {
	ID   string
	Name string
}

type onboardingRequestBody struct {
	Onboarding struct {
		InstitutionName    string `json:"institution_name,omitempty"`
		InstitutionZipCode string `json:"institution_zip_code,omitempty"`
		InstitutionWebsite string `json:"institution_website,omitempty"`
		InstitutionEmail   string `json:"institution_email,omitempty"`
		Subdomain          string `json:"subdomain,omitempty"`
		EHRName            string `json:"ehr_name,omitempty"`
	} `json:"onboarding"`
}

type onboardingResponse struct {
	Code bool `json:"code"`
	Data struct {
		ID            string    `json:"id"`
		Subdomain     string    `json:"subdomain"`
		URL           string    `json:"url"`
		URLExpiresAt  time.Time `json:"url_expires_at"`
		Status        string    `json:"status"`
		BookingParams struct {
			LocationID string `json:"location_id"`
		}
	} `json:"data"`
}

type patientRequestBody struct {
	Provider struct {
		ProviderID int `json:"provider_id"`
	} `json:"provider"`
	Patient struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Bio       struct {
			DateOfBirth string `json:"date_of_birth"`
			PhoneNumber string `json:"phone_number"`
		} `json:"bio"`
	} `json:"patient"`
	ReturnExistingIfMatch bool `json:"return_existing_if_match"`
}

type patientResponse struct {
	Code bool `json:"code"`
	Data struct {
		User struct {
			ID int `json:"id"`
		} `json:"user"`
	} `json:"data"`
}

type patientsResponse struct {
	Code bool `json:"code"`
	Data []struct {
		ID int `json:"id"`
	} `json:"data"`
}

type appointmentRequestBody struct {
	Appt struct {
		PatientID  int    `json:"patient_id"`
		ProviderID int    `json:"provider_id"`
		StartTime  string `json:"start_time"`
		Note       string `json:"note,omitempty"`
	} `json:"appt"`
}

type nexAppointment struct {
	ID         int64     `json:"id"`
	ProviderID int       `json:"provider_id"`
	StartTime  time.Time `json:"start_time"`
	Confirmed  bool      `json:"confirmed"`
	Cancelled  bool      `json:"cancelled"`
}

type appointmentResponse struct {
	Code bool `json:"code"`
	Data struct {
		Appt nexAppointment `json:"appt"`
	} `json:"data"`
}

type appointmentsResponse struct {
	Code bool             `json:"code"`
	Data []nexAppointment `json:"data"`
}

type NexHealth struct {
	http        *resty.Client
	apiKey      string
	dbMux       storage.IDBMux
	authStorage storage.INexHealthAuthStorage

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

var _ ehr.IEHR = (*NexHealth)(nil)

func New(cfg internal.NexHealthConfig, dbMux storage.IDBMux, authStorage storage.INexHealthAuthStorage) *NexHealth {
	client := resty.New().
		SetBaseURL(strings.TrimRight(cfg.BaseURL, "/")).
		SetTimeout(requestTimeout).
		SetHeader("Accept", acceptHeader)

	return &NexHealth{
		http:        client,
		apiKey:      cfg.APIKey,
		dbMux:       dbMux,
		authStorage: authStorage,
	}
}

func (n *NexHealth) Authenticate(ctx context.Context) error {
	_, err := n.authenticate(ctx)
	return err
}

func (n *NexHealth) FindAppointment(ctx context.Context, params ehr.AvailableAppointmentsParams) ([]ehr.Appointment, error) {
	providers, err := n.listProviders(ctx, params.Subdomain, params.LocationID)
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return nil, nil
	}

	providerNames := make(map[string]string, len(providers))
	for _, p := range providers {
		providerNames[p.ID] = p.Name
	}

	token, err := n.bearerToken(ctx)
	if err != nil {
		return nil, err
	}

	startDate := params.StartDate
	if startDate.IsZero() {
		startDate = time.Now()
	}
	days := params.Days
	if days <= 0 {
		days = defaultSearchDays
	}

	query := url.Values{}
	query.Set("subdomain", params.Subdomain)
	query.Set("start_date", startDate.Format("2006-01-02"))
	query.Set("days", strconv.Itoa(days))
	query.Add("lids[]", params.LocationID)
	for _, p := range providers {
		query.Add("pids[]", p.ID)
	}

	var out appointmentSlotsResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetQueryParamsFromValues(query).
		SetResult(&out).
		Get("/appointment_slots")
	if err != nil {
		return nil, rerror.New(fmt.Errorf("nexhealth: find appointment: %w", err))
	}
	if resp.IsError() {
		return nil, rerror.New(fmt.Errorf("nexhealth: find appointment: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	out.setProviderNames(providerNames)
	return out.toAppointments()
}

func (n *NexHealth) CreateOnboarding(ctx context.Context, params ehr.CreateOnboardingParams) (*ehr.Onboarding, error) {
	token, err := n.bearerToken(ctx)
	if err != nil {
		return nil, err
	}

	var body onboardingRequestBody
	body.Onboarding.InstitutionName = params.InstitutionName
	body.Onboarding.InstitutionZipCode = params.InstitutionZipCode
	body.Onboarding.InstitutionWebsite = params.InstitutionWebsite
	body.Onboarding.InstitutionEmail = params.InstitutionEmail
	body.Onboarding.Subdomain = params.Subdomain
	body.Onboarding.EHRName = params.EHRName

	var out onboardingResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetHeader("Nex-Api-Version", onboardingsAPIVersion).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&out).
		Post("/onboardings")
	if err != nil {
		return nil, rerror.New(fmt.Errorf("nexhealth: create onboarding: %w", err))
	}
	if resp.IsError() {
		return nil, rerror.New(fmt.Errorf("nexhealth: create onboarding: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	return &ehr.Onboarding{
		ID:           out.Data.ID,
		Subdomain:    out.Data.Subdomain,
		URL:          out.Data.URL,
		URLExpiresAt: out.Data.URLExpiresAt,
		Status:       out.Data.Status,
		LocationID:   out.Data.BookingParams.LocationID,
	}, nil
}

func (n *NexHealth) BookAppointment(ctx context.Context, params ehr.BookAppointmentParams) (*ehr.BookedAppointment, error) {
	providerID, err := strconv.Atoi(params.ProviderID)
	if err != nil {
		return nil, rerror.New(fmt.Errorf("nexhealth: book appointment: invalid provider id %q: %w", params.ProviderID, err))
	}

	patientID, err := n.createOrMatchPatient(ctx, params, providerID)
	if err != nil {
		return nil, err
	}

	token, err := n.bearerToken(ctx)
	if err != nil {
		return nil, err
	}

	var body appointmentRequestBody
	body.Appt.PatientID = patientID
	body.Appt.ProviderID = providerID
	body.Appt.StartTime = params.Time.Format(time.RFC3339)
	body.Appt.Note = params.Note

	query := url.Values{}
	query.Set("subdomain", params.Subdomain)
	query.Set("location_id", params.LocationID)

	var out appointmentResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetHeader("Nex-Api-Version", appointmentsAPIVersion).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetQueryParamsFromValues(query).
		SetBody(body).
		SetResult(&out).
		Post("/appointments")
	if err != nil {
		return nil, rerror.New(fmt.Errorf("nexhealth: book appointment: %w", err))
	}
	if resp.IsError() {
		return nil, rerror.New(fmt.Errorf("nexhealth: book appointment: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	appt := out.Data.Appt
	bookedTime := appt.StartTime
	if bookedTime.IsZero() {
		bookedTime = params.Time
	}
	bookedProviderID := appt.ProviderID
	if bookedProviderID == 0 {
		bookedProviderID = providerID
	}
	return &ehr.BookedAppointment{
		ID:         strconv.FormatInt(appt.ID, 10),
		Time:       bookedTime,
		ProviderID: strconv.Itoa(bookedProviderID),
		Confirmed:  appt.Confirmed,
	}, nil
}

func (n *NexHealth) CancelAppointment(ctx context.Context, params ehr.CancelAppointmentParams) (*ehr.CancelledAppointment, error) {
	patientID, err := n.findPatient(ctx, params.Subdomain, params.LocationID, params.Patient)
	if err != nil {
		return nil, err
	}

	appt, err := n.nextUpcomingAppointment(ctx, params.Subdomain, params.LocationID, patientID)
	if err != nil {
		return nil, err
	}

	if err := n.cancelAppointment(ctx, params.Subdomain, appt.ID); err != nil {
		return nil, err
	}

	return &ehr.CancelledAppointment{
		ID:         strconv.FormatInt(appt.ID, 10),
		Time:       appt.StartTime,
		ProviderID: strconv.Itoa(appt.ProviderID),
	}, nil
}

func (n *NexHealth) createOrMatchPatient(ctx context.Context, params ehr.BookAppointmentParams, providerID int) (int, error) {
	token, err := n.bearerToken(ctx)
	if err != nil {
		return 0, err
	}

	var body patientRequestBody
	body.Provider.ProviderID = providerID
	body.Patient.FirstName = params.Patient.FirstName
	body.Patient.LastName = params.Patient.LastName
	body.Patient.Email = params.Patient.Email
	body.Patient.Bio.DateOfBirth = params.Patient.DateOfBirth
	body.Patient.Bio.PhoneNumber = params.Patient.Phone
	body.ReturnExistingIfMatch = true

	query := url.Values{}
	query.Set("subdomain", params.Subdomain)
	query.Set("location_id", params.LocationID)

	var out patientResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetHeader("Nex-Api-Version", appointmentsAPIVersion).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetQueryParamsFromValues(query).
		SetBody(body).
		SetResult(&out).
		Post("/patients")
	if err != nil {
		return 0, rerror.New(fmt.Errorf("nexhealth: create patient: %w", err))
	}
	if resp.IsError() {
		return 0, rerror.New(fmt.Errorf("nexhealth: create patient: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}
	if out.Data.User.ID == 0 {
		return 0, rerror.New(fmt.Errorf("nexhealth: create patient: empty patient id in response: %s", resp.String()))
	}
	return out.Data.User.ID, nil
}

func (n *NexHealth) findPatient(ctx context.Context, subdomain, locationID string, lookup ehr.PatientLookup) (int, error) {
	token, err := n.bearerToken(ctx)
	if err != nil {
		return 0, err
	}

	query := url.Values{}
	query.Set("subdomain", subdomain)
	query.Set("location_id", locationID)
	if lookup.Name != "" {
		query.Set("name", lookup.Name)
	}
	if lookup.Phone != "" {
		query.Set("phone_number", lookup.Phone)
	}
	if lookup.DateOfBirth != "" {
		query.Set("date_of_birth", lookup.DateOfBirth)
	}

	var out patientsResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetHeader("Nex-Api-Version", appointmentsAPIVersion).
		SetHeader("Accept", "application/json").
		SetQueryParamsFromValues(query).
		SetResult(&out).
		Get("/patients")
	if err != nil {
		return 0, rerror.New(fmt.Errorf("nexhealth: find patient: %w", err))
	}
	if resp.IsError() {
		return 0, rerror.New(fmt.Errorf("nexhealth: find patient: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	switch len(out.Data) {
	case 0:
		return 0, rerror.NewMessage("no patient matched the provided details", rerror.Validation)
	case 1:
		return out.Data[0].ID, nil
	default:
		return 0, rerror.NewMessage(fmt.Sprintf("%d patients matched the provided details, need more to identify one", len(out.Data)), rerror.Validation)
	}
}

func (n *NexHealth) nextUpcomingAppointment(ctx context.Context, subdomain, locationID string, patientID int) (nexAppointment, error) {
	token, err := n.bearerToken(ctx)
	if err != nil {
		return nexAppointment{}, err
	}

	now := time.Now()
	query := url.Values{}
	query.Set("subdomain", subdomain)
	query.Set("location_id", locationID)
	query.Add("patient_ids[]", strconv.Itoa(patientID))
	query.Set("start", now.Format(time.RFC3339))
	query.Set("end", now.Add(cancelSearchWindow).Format(time.RFC3339))
	query.Set("cancelled", "false")
	query.Set("per_page", strconv.Itoa(appointmentsPageSize))

	var out appointmentsResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetHeader("Nex-Api-Version", appointmentsAPIVersion).
		SetHeader("Accept", "application/json").
		SetQueryParamsFromValues(query).
		SetResult(&out).
		Get("/appointments")
	if err != nil {
		return nexAppointment{}, rerror.New(fmt.Errorf("nexhealth: list appointments: %w", err))
	}
	if resp.IsError() {
		return nexAppointment{}, rerror.New(fmt.Errorf("nexhealth: list appointments: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	soonest := -1
	for i := range out.Data {
		if out.Data[i].Cancelled {
			continue
		}
		if soonest == -1 || out.Data[i].StartTime.Before(out.Data[soonest].StartTime) {
			soonest = i
		}
	}
	if soonest == -1 {
		return nexAppointment{}, rerror.NewMessage("no upcoming appointment found for the patient", rerror.Validation)
	}
	return out.Data[soonest], nil
}

func (n *NexHealth) cancelAppointment(ctx context.Context, subdomain string, appointmentID int64) error {
	token, err := n.bearerToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]any{"appt": map[string]any{"cancelled": true}}

	query := url.Values{}
	query.Set("subdomain", subdomain)

	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetHeader("Nex-Api-Version", appointmentsAPIVersion).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetQueryParamsFromValues(query).
		SetBody(body).
		Patch("/appointments/" + strconv.FormatInt(appointmentID, 10))
	if err != nil {
		return rerror.New(fmt.Errorf("nexhealth: cancel appointment: %w", err))
	}
	if resp.IsError() {
		return rerror.New(fmt.Errorf("nexhealth: cancel appointment: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}
	return nil
}

func (n *NexHealth) listProviders(ctx context.Context, subdomain, locationID string) ([]nexProvider, error) {
	token, err := n.bearerToken(ctx)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("subdomain", subdomain)
	if locationID != "" {
		query.Set("location_id", locationID)
	}

	var out providersResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetQueryParamsFromValues(query).
		SetResult(&out).
		Get("/providers")
	if err != nil {
		return nil, rerror.New(fmt.Errorf("nexhealth: list providers: %w", err))
	}
	if resp.IsError() {
		return nil, rerror.New(fmt.Errorf("nexhealth: list providers: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	providers := make([]nexProvider, 0, len(out.Data))
	for _, p := range out.Data {
		if p.Inactive {
			continue
		}
		name := p.DisplayName
		if name == "" {
			name = p.Name
		}
		providers = append(providers, nexProvider{ID: strconv.Itoa(p.ID), Name: name})
	}
	return providers, nil
}

func (n *NexHealth) bearerToken(ctx context.Context) (string, error) {
	n.mu.Lock()
	if n.token != "" && time.Now().Add(tokenSafetyMargin).Before(n.tokenExpiry) {
		token := n.token
		n.mu.Unlock()
		return token, nil
	}
	n.mu.Unlock()
	return n.authenticate(ctx)
}

func (n *NexHealth) authenticate(ctx context.Context) (string, error) {
	if n.apiKey == "" {
		return "", rerror.New(fmt.Errorf("nexhealth: API key not configured (set nexhealth.api_key in config)"))
	}

	db, err := n.dbMux.BeginWrite()
	if err != nil {
		return "", rerror.New(fmt.Errorf("nexhealth: acquire db writer: %w", err))
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return "", rerror.New(fmt.Errorf("nexhealth: begin auth transaction: %w", err))
	}
	defer tx.Rollback(ctx)

	record, err := n.authStorage.LockForUpdate(ctx, tx)
	if err != nil {
		return "", rerror.New(fmt.Errorf("nexhealth: lock auth token: %w", err))
	}

	if record.Token != "" && time.Now().Add(tokenSafetyMargin).Before(record.ExpiresAt) {
		if err := tx.Commit(ctx); err != nil {
			return "", rerror.New(fmt.Errorf("nexhealth: commit auth transaction: %w", err))
		}
		n.cacheToken(record.Token, record.ExpiresAt)
		return record.Token, nil
	}

	token, expiry, err := n.requestToken(ctx)
	if err != nil {
		return "", err
	}

	if err := n.authStorage.Update(ctx, tx, token, expiry); err != nil {
		return "", rerror.New(fmt.Errorf("nexhealth: store auth token: %w", err))
	}
	if err := tx.Commit(ctx); err != nil {
		return "", rerror.New(fmt.Errorf("nexhealth: commit auth transaction: %w", err))
	}

	n.cacheToken(token, expiry)
	return token, nil
}

func (n *NexHealth) requestToken(ctx context.Context) (string, time.Time, error) {
	var out authResponse
	resp, err := n.http.R().
		SetContext(ctx).
		SetHeader("Authorization", n.apiKey).
		SetResult(&out).
		Post("/authenticates")
	if err != nil {
		return "", time.Time{}, rerror.New(fmt.Errorf("nexhealth: authenticate: %w", err))
	}
	if resp.IsError() {
		return "", time.Time{}, rerror.New(fmt.Errorf("nexhealth: authenticate: unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}
	if out.Data.Token == "" {
		return "", time.Time{}, rerror.New(fmt.Errorf("nexhealth: authenticate: empty token in response: %s", resp.String()))
	}

	expiry := tokenExpiryFromJWT(out.Data.Token)
	if expiry.IsZero() {
		expiry = time.Now().Add(fallbackTokenTTL)
	}
	return out.Data.Token, expiry, nil
}

func (n *NexHealth) cacheToken(token string, expiry time.Time) {
	n.mu.Lock()
	n.token = token
	n.tokenExpiry = expiry
	n.mu.Unlock()
}

func tokenExpiryFromJWT(token string) time.Time {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return time.Time{}
	}
	return time.Unix(claims.Exp, 0)
}

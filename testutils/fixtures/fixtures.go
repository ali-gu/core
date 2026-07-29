package fixtures

import (
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type practiceParams struct {
	Name    string
	Email   *string
	ZipCode *string
	Website *string
}

type PracticeOption func(*practiceParams)

func WithPracticeName(name string) PracticeOption {
	return func(p *practiceParams) { p.Name = name }
}

func WithPracticeEmail(email string) PracticeOption {
	return func(p *practiceParams) { p.Email = &email }
}

func WithPracticeZipCode(zipCode string) PracticeOption {
	return func(p *practiceParams) { p.ZipCode = &zipCode }
}

func WithPracticeWebsite(website string) PracticeOption {
	return func(p *practiceParams) { p.Website = &website }
}

func WithPracticeNoContactInfo() PracticeOption {
	return func(p *practiceParams) {
		p.Email = nil
		p.ZipCode = nil
		p.Website = nil
	}
}

func NewPractice(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...PracticeOption) storage.Practice {
	t.Helper()

	params := practiceParams{
		Name:    "foo_practice",
		Email:   ptr.To("ops@example.com"),
		ZipCode: ptr.To("62701"),
		Website: ptr.To("www.example.com"),
	}
	for _, opt := range opts {
		opt(&params)
	}

	practice, err := bz.Practice.Create(cfg.Ctx, cfg.DB, contracts.CreatePracticeRequest{
		Name:    params.Name,
		Email:   params.Email,
		ZipCode: params.ZipCode,
		Website: params.Website,
	})
	require.NoError(t, err)

	return *practice
}

type locationParams struct {
	PracticeID *ksuid.KSUID
	Address    string
	EHR        constants.NexHealthEHR
}

type LocationOption func(*locationParams)

func WithLocationPracticeID(id ksuid.KSUID) LocationOption {
	return func(p *locationParams) { p.PracticeID = &id }
}

func WithLocationAddress(address string) LocationOption {
	return func(p *locationParams) { p.Address = address }
}

func WithLocationEHR(ehrType constants.NexHealthEHR) LocationOption {
	return func(p *locationParams) { p.EHR = ehrType }
}

func defaultLocationParams() locationParams {
	return locationParams{
		Address: "100 Main St",
		EHR:     constants.NexHealthEHROpenDental,
	}
}

func NewLocation(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...LocationOption) storage.Location {
	t.Helper()

	params := defaultLocationParams()
	for _, opt := range opts {
		opt(&params)
	}

	practiceID := params.PracticeID
	if practiceID == nil {
		practice := NewPractice(t, cfg, bz)
		practiceID = &practice.ID
	}

	ehrMock := cfg.Deps.EHR.(*ehr.MockIEHR)
	ehrMock.On("CreateOnboarding", mock.Anything, mock.Anything).Return(&ehr.Onboarding{
		ID:        "onboarding_id",
		Subdomain: "default-subdomain",
		Status:    "in_progress",
	}, nil).Once()

	location, _, err := bz.Location.Create(cfg.Ctx, cfg.DB, *practiceID, contracts.CreateLocationRequest{
		Address: params.Address,
		EHR:     params.EHR,
	})
	require.NoError(t, err)
	return *location
}

func NewPendingLocation(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...LocationOption) storage.Location {
	t.Helper()

	params := defaultLocationParams()
	for _, opt := range opts {
		opt(&params)
	}

	practiceID := params.PracticeID
	if practiceID == nil {
		practice := NewPractice(t, cfg, bz)
		practiceID = &practice.ID
	}

	location := storage.Location{
		EntityBase: storage.EntityBase[states.LocationState]{EntityState: states.LocationStatePending},
		ID:         ksuid.New(),
		Address:    params.Address,
		PracticeID: *practiceID,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, cfg.Deps.Storage.Location.Create(cfg.Ctx, cfg.DB, location))
	return location
}

type agentParams struct {
	PracticeID    *ksuid.KSUID
	Name          string
	LocationID    *ksuid.KSUID
	PhoneNumberID *ksuid.KSUID
	AgentRef      *string
}

type AgentOption func(*agentParams)

func WithAgentPracticeID(id ksuid.KSUID) AgentOption {
	return func(p *agentParams) { p.PracticeID = &id }
}

func WithAgentName(name string) AgentOption {
	return func(p *agentParams) { p.Name = name }
}

func WithAgentLocationID(id ksuid.KSUID) AgentOption {
	return func(p *agentParams) { p.LocationID = &id }
}

func WithAgentPhoneNumberID(id ksuid.KSUID) AgentOption {
	return func(p *agentParams) { p.PhoneNumberID = &id }
}

func WithAgentRef(ref string) AgentOption {
	return func(p *agentParams) { p.AgentRef = &ref }
}

func NewAgent(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...AgentOption) storage.Agent {
	t.Helper()

	params := agentParams{Name: "foo_agent"}
	for _, opt := range opts {
		opt(&params)
	}

	practiceID := params.PracticeID
	if practiceID == nil {
		if params.LocationID != nil {
			location, err := cfg.Deps.Storage.Location.GetByID(cfg.Ctx, cfg.DB, *params.LocationID)
			require.NoError(t, err)
			practiceID = &location.PracticeID
		} else {
			practice := NewPractice(t, cfg, bz)
			practiceID = &practice.ID
		}
	}

	agentRecord, err := bz.Agent.Create(cfg.Ctx, cfg.DB, *practiceID, contracts.CreateAgentRequest{
		Name:       params.Name,
		LocationID: params.LocationID,
	})
	require.NoError(t, err)

	if params.PhoneNumberID != nil {
		agentRecord, err = bz.Agent.Update(cfg.Ctx, cfg.DB, agentRecord.ID, contracts.UpdateAgentRequest{
			PhoneNumberID: params.PhoneNumberID,
		})
		require.NoError(t, err)
	}

	if params.AgentRef != nil {
		agentRecord.AgentRef = params.AgentRef
		require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, *agentRecord))
		agentRecord, err = cfg.Deps.Storage.Agent.GetByID(cfg.Ctx, cfg.DB, agentRecord.ID)
		require.NoError(t, err)
	}

	return *agentRecord
}

type ehrParams struct {
	LocationID  *ksuid.KSUID
	Type        constants.EHR
	Subdomain   string
	LocationRef *string
}

type EHROption func(*ehrParams)

func WithEHRLocationID(id ksuid.KSUID) EHROption {
	return func(p *ehrParams) { p.LocationID = &id }
}

func WithEHRType(ehrType constants.EHR) EHROption {
	return func(p *ehrParams) { p.Type = ehrType }
}

func WithEHRSubdomain(subdomain string) EHROption {
	return func(p *ehrParams) { p.Subdomain = subdomain }
}

func WithEHRLocationRef(ref string) EHROption {
	return func(p *ehrParams) { p.LocationRef = &ref }
}

func NewEHR(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...EHROption) storage.EHRS {
	t.Helper()

	params := ehrParams{
		Type:        constants.EHRNexHealth,
		Subdomain:   "acme-dental",
		LocationRef: ptr.To("nexhealth_location_id"),
	}
	for _, opt := range opts {
		opt(&params)
	}

	locationID := params.LocationID
	if locationID == nil {
		location := NewPendingLocation(t, cfg, bz)
		locationID = &location.ID
	}

	ehrRecord := storage.EHRS{
		ID:            ksuid.New(),
		Type:          params.Type,
		Subdomain:     params.Subdomain,
		LocationRef:   params.LocationRef,
		LocationID:    *locationID,
		OnboardingURL: "https://app.nexhealth.com/onboardings/onboarding_id",
		OnboardingID:  "onboarding_id",
		CreatedAt:     time.Now(),
	}
	require.NoError(t, cfg.Deps.Storage.EHR.Create(cfg.Ctx, cfg.DB, ehrRecord))

	location, err := cfg.Deps.Storage.Location.GetByID(cfg.Ctx, cfg.DB, *locationID)
	require.NoError(t, err)
	location.LocationToActive()
	require.NoError(t, cfg.Deps.Storage.Location.Update(cfg.Ctx, cfg.DB, *location))

	return ehrRecord
}

type phoneNumberParams struct {
	LocationID       *ksuid.KSUID
	PhoneNumber      string
	PhoneNumberIDRef *string
}

type PhoneNumberOption func(*phoneNumberParams)

func WithPhoneNumberLocationID(id ksuid.KSUID) PhoneNumberOption {
	return func(p *phoneNumberParams) { p.LocationID = &id }
}

func WithPhoneNumberNumber(number string) PhoneNumberOption {
	return func(p *phoneNumberParams) { p.PhoneNumber = number }
}

func WithPhoneNumberIDRef(ref string) PhoneNumberOption {
	return func(p *phoneNumberParams) { p.PhoneNumberIDRef = &ref }
}

func WithPhoneNumberNoIDRef() PhoneNumberOption {
	return func(p *phoneNumberParams) { p.PhoneNumberIDRef = nil }
}

func NewPhoneNumber(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...PhoneNumberOption) storage.PhoneNumber {
	t.Helper()

	params := phoneNumberParams{PhoneNumber: "+15555550100", PhoneNumberIDRef: ptr.To("telnyx_number_ref")}
	for _, opt := range opts {
		opt(&params)
	}

	var locationID ksuid.KSUID
	if params.LocationID == nil {
		location := NewLocation(t, cfg, bz)
		locationID = location.ID
	} else {
		locationID = *params.LocationID
	}
	location, err := bz.Location.GetByID(cfg.Ctx, cfg.DB, locationID)
	require.NoError(t, err)

	phoneNumber := storage.PhoneNumber{
		EntityBase:       storage.EntityBase[states.PhoneNumberState]{EntityState: states.PhoneNumberStateActive},
		ID:               ksuid.New(),
		PhoneNumberIDRef: params.PhoneNumberIDRef,
		PhoneNumber:      params.PhoneNumber,
		PracticeID:       location.PracticeID,
		CreatedAt:        time.Now(),
	}
	require.NoError(t, cfg.Deps.Storage.PhoneNumber.Create(cfg.Ctx, cfg.DB, phoneNumber))

	agents, err := cfg.Deps.Storage.Agent.Get(cfg.Ctx, cfg.DB, storage.AgentFilters{LocationID: &locationID})
	require.NoError(t, err)
	if len(agents) > 0 {
		agentRecord := agents[0]
		agentRecord.AssignPhoneNumber(phoneNumber.ID)
		require.NoError(t, cfg.Deps.Storage.Agent.Update(cfg.Ctx, cfg.DB, agentRecord))
	}

	return phoneNumber
}

type toolParams struct {
	Type     constants.ToolType
	Kind     constants.ToolKind
	ToolRef  string
	Config   map[string]any
	Disabled bool
}

type ToolOption func(*toolParams)

func WithToolType(toolType constants.ToolType) ToolOption {
	return func(p *toolParams) { p.Type = toolType }
}

func WithToolKind(kind constants.ToolKind) ToolOption {
	return func(p *toolParams) { p.Kind = kind }
}

func WithToolRef(ref string) ToolOption {
	return func(p *toolParams) { p.ToolRef = ref }
}

func WithToolConfig(config map[string]any) ToolOption {
	return func(p *toolParams) { p.Config = config }
}

func WithToolDisabled() ToolOption {
	return func(p *toolParams) { p.Disabled = true }
}

func NewTool(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz, opts ...ToolOption) storage.Tool {
	t.Helper()

	params := toolParams{
		Type:    constants.ToolTypeWebhook,
		Kind:    constants.ToolKindBookAppointment,
		ToolRef: "tool_ref_" + ksuid.New().String(),
		Config:  map[string]any{},
	}
	for _, opt := range opts {
		opt(&params)
	}

	toolRecord := storage.Tool{
		EntityBase: storage.EntityBase[states.ToolState]{EntityState: states.ToolStateActive},
		ID:         ksuid.New(),
		Type:       params.Type,
		Kind:       params.Kind,
		ToolRef:    params.ToolRef,
		Config:     params.Config,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, cfg.Deps.Storage.Tool.Create(cfg.Ctx, cfg.DB, toolRecord))

	if params.Disabled {
		toolRecord.ToolToDisabled()
		require.NoError(t, cfg.Deps.Storage.Tool.Update(cfg.Ctx, cfg.DB, toolRecord))
	}

	found, err := cfg.Deps.Storage.Tool.GetByID(cfg.Ctx, cfg.DB, toolRecord.ID)
	require.NoError(t, err)
	return *found
}

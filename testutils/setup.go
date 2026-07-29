package testutils

import (
	"context"
	"testing"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/services/tool"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const TestDomain = "https://test.example.com"

type TestConfig struct {
	Ctx  context.Context
	Cfg  internal.Config
	DB   storage.DB
	Mux  *storage.DBMux
	Deps biz.Dependencies
}

func BasicSetup(t *testing.T) (*TestConfig, biz.Biz) {
	t.Helper()

	cfg, bz := Setup(t)
	t.Cleanup(func() {
		CleanupDB(t, *cfg)
	})
	return cfg, bz
}

func CleanupDB(t *testing.T, cfg TestConfig) {
	t.Helper()

	require.NoError(t, cfg.DB.(storage.DBTx).Rollback(cfg.Ctx))
	cfg.Mux.Close()
}

func Setup(t *testing.T) (*TestConfig, biz.Biz) {
	t.Helper()

	deps, bz := SetupBiz(t)
	return setupTx(t, deps), bz
}

func BasicSetupWithDeps(t *testing.T, deps biz.Dependencies) (*TestConfig, biz.Biz) {
	t.Helper()

	cfg, bz := SetupWithDeps(t, deps)
	t.Cleanup(func() {
		CleanupDB(t, *cfg)
	})
	return cfg, bz
}

func SetupWithDeps(t *testing.T, deps biz.Dependencies) (*TestConfig, biz.Biz) {
	t.Helper()

	return setupTx(t, deps), *biz.NewBiz(deps)
}

func setupTx(t *testing.T, deps biz.Dependencies) *TestConfig {
	t.Helper()

	ctx := context.Background()
	cfg, err := internal.NewConfig(constants.LandscapeTest)
	require.NoError(t, err)

	mux, err := storage.NewDBMux(ctx, *cfg)
	require.NoError(t, err)

	db, err := mux.BeginWrite()
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	return &TestConfig{
		Ctx:  ctx,
		Cfg:  *cfg,
		DB:   tx,
		Mux:  mux,
		Deps: deps,
	}
}

func SetupBiz(t *testing.T) (biz.Dependencies, biz.Biz) {
	agentMock := agent.NewMockIAgent(t)
	phoneNumberMock := phonenumber.NewMockIPhoneNumberManager(t)
	ehrMock := ehr.NewMockIEHR(t)
	authMock := auth.NewMockIAuth(t)
	toolMock := tool.NewMockITool(t)

	setupDefaultAgentMocks(agentMock)
	setupDefaultPhoneNumberMocks(phoneNumberMock)
	setupDefaultEHRMocks(ehrMock)
	setupDefaultAuthMocks(authMock)
	setupDefaultToolMocks(toolMock)

	dep := biz.Dependencies{
		Storage:                  storage.NewStorage(),
		TelnyxAgent:              agentMock,
		TelnyxPhoneNumberManager: phoneNumberMock,
		SupabaseAuth:             authMock,
		EHR:                      ehrMock,
		TelnyxTool:               toolMock,
		Domain:                   TestDomain,
	}

	return dep, *biz.NewBiz(dep)
}

func setupDefaultAgentMocks(agentMock *agent.MockIAgent) {
	agentMock.On("Create", mock.Anything, mock.MatchedBy(func(params agent.CreateAgentParams) bool {
		return params.Name != ""
	})).Return(&agent.CreateAgentResult{
		ID:   "telnyx_agent_id",
		Name: "telnyx_agent_name",
	}, nil).Maybe()
	agentMock.On("Delete", mock.Anything, mock.Anything).Return(nil).Maybe()
	agentMock.On("GetAnalytics", mock.Anything, mock.Anything).Return([]agent.ConversationAnalytics{}, nil).Maybe()
	agentMock.On("GetConversation", mock.Anything, mock.Anything).Return(&agent.ConversationAnalytics{}, nil).Maybe()
	agentMock.On("GetRecordings", mock.Anything, mock.Anything).Return([]agent.ConversationRecording{}, nil).Maybe()
}

func setupDefaultPhoneNumberMocks(phoneNumberMock *phonenumber.MockIPhoneNumberManager) {
	phoneNumberMock.On("Reserve", mock.Anything, mock.Anything).Return(&phonenumber.ReservePhoneNumberResult{
		ReservationRef: "reservation_ref",
	}, nil).Maybe()
}

func setupDefaultEHRMocks(ehrMock *ehr.MockIEHR) {
}

func setupDefaultAuthMocks(supabaseMock *auth.MockIAuth) {
	supabaseMock.On("SignUp", mock.Anything, mock.Anything, mock.Anything).Return(&auth.SignUpResult{
		ID: "supabase_user_ref",
	}, nil).Maybe()
}

func setupDefaultToolMocks(toolMock *tool.MockITool) {
	toolMock.On("List", mock.Anything).Return([]tool.ListToolsResult{}, nil).Maybe()
}

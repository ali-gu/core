package api_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/api"
	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/testutils"
	"github.com/ali-gulzar/speechory-core/testutils/fixtures"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/team-telnyx/telnyx-go/v4/lib"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type txMux struct {
	db storage.DB
}

func (m txMux) BeginWrite() (storage.DB, error)  { return m.db, nil }
func (m txMux) BeginRead() (storage.DB, error)   { return m.db, nil }
func (m txMux) PingWriter(context.Context) error { return nil }
func (m txMux) PingReader(context.Context) error { return nil }
func (m txMux) Close()                           {}

type e2e struct {
	t                *testing.T
	cfg              *testutils.TestConfig
	bz               biz.Biz
	router           *gin.Engine
	authMock         *auth.MockIAuth
	telnyxPrivateKey ed25519.PrivateKey
}

func newE2E(t *testing.T) *e2e {
	t.Helper()

	cfg, bz := testutils.BasicSetup(t)
	return newE2EFrom(t, cfg, bz)
}

func newE2EWithDeps(t *testing.T, deps biz.Dependencies) *e2e {
	t.Helper()

	cfg, bz := testutils.BasicSetupWithDeps(t, deps)
	return newE2EFrom(t, cfg, bz)
}

func newE2EFrom(t *testing.T, cfg *testutils.TestConfig, bz biz.Biz) *e2e {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	a := &api.API{
		Biz:   &bz,
		DBMux: txMux{cfg.DB},
		Config: internal.Config{
			Telnyx: internal.TelnyxConfig{PublicKey: base64.StdEncoding.EncodeToString(pub)},
		},
		Storage: storage.NewStorage(),
	}

	return &e2e{
		t:                t,
		cfg:              cfg,
		bz:               bz,
		router:           a.Routes(),
		authMock:         cfg.Deps.SupabaseAuth.(*auth.MockIAuth),
		telnyxPrivateKey: priv,
	}
}

func (e *e2e) do(method, path, token string, body any) *httptest.ResponseRecorder {
	e.t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(e.t, err)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func (e *e2e) doWithCookie(method, path, token string, cookie *http.Cookie, body any) *httptest.ResponseRecorder {
	e.t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(e.t, err)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func responseCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func (e *e2e) doWebhook(method, path string, body any) *httptest.ResponseRecorder {
	return e.doWebhookWithHeaders(method, path, nil, body)
}

func (e *e2e) doWebhookWithHeaders(method, path string, headers map[string]string, body any) *httptest.ResponseRecorder {
	e.t.Helper()

	var raw []byte
	var reader io.Reader
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		require.NoError(e.t, err)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := ed25519.Sign(e.telnyxPrivateKey, []byte(timestamp+"|"+string(raw)))
	req.Header.Set(lib.TimestampHeader, timestamp)
	req.Header.Set(lib.SignatureHeader, base64.StdEncoding.EncodeToString(signature))

	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func (e *e2e) seedRole(roleType constants.RoleType) storage.Role {
	e.t.Helper()

	roles, err := e.cfg.Deps.Storage.Role.Get(e.cfg.Ctx, e.cfg.DB, storage.RoleFilters{Type: &roleType})
	require.NoError(e.t, err)
	require.Len(e.t, roles, 1)
	return roles[0]
}

func (e *e2e) clearRoleType(roleType constants.RoleType) {
	e.t.Helper()

	_, err := e.cfg.DB.Exec(e.cfg.Ctx, "DELETE FROM roles WHERE type = $1", roleType)
	require.NoError(e.t, err)
}

func (e *e2e) authFor(practiceID ksuid.KSUID) string {
	e.t.Helper()

	return e.authForRole(practiceID, constants.RoleTypeAdmin)
}

func (e *e2e) authForRole(practiceID ksuid.KSUID, roleType constants.RoleType) string {
	e.t.Helper()

	token := "token_" + ksuid.New().String()
	ref := "ref_" + ksuid.New().String()

	role := e.seedRole(roleType)
	require.NoError(e.t, e.cfg.Deps.Storage.User.Create(e.cfg.Ctx, e.cfg.DB, storage.User{
		EntityBase: storage.EntityBase[states.UserState]{EntityState: states.UserStateActive},
		ID:         ksuid.New(),
		UserRef:    ref,
		RoleID:     role.ID,
		PracticeID: practiceID,
		Email:      ref + "@example.com",
		CreatedAt:  time.Now(),
	}))

	e.setupAuthMock(token, ref)

	return token
}

func (e *e2e) setupAuthMock(token, ref string) {
	e.authMock.On("Authenticate", mock.Anything, token).
		Return(&auth.AuthenticatedUser{ID: ref, Email: ref + "@example.com"}, nil).
		Maybe()
}

func (e *e2e) authedPractice() (storage.Practice, string) {
	e.t.Helper()

	practice := fixtures.NewPractice(e.t, e.cfg, e.bz)
	return practice, e.authFor(practice.ID)
}

func (e *e2e) authedSuperAdminPractice() (storage.Practice, string) {
	e.t.Helper()

	practice := fixtures.NewPractice(e.t, e.cfg, e.bz)
	return practice, e.authForRole(practice.ID, constants.RoleTypeSuperAdmin)
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, target any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), target))
}

package authsvc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jarimnark/sugar-cane-service/internal/config"
	authsvc "github.com/Jarimnark/sugar-cane-service/internal/service/auth"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

// mockUserSvc is a test double for usersvc.Service.
type mockUserSvc struct {
	loginWithOAuthFn func(ctx context.Context, provider, providerID, email, displayName, photoURL, ip, userAgent string) (*userdto.LoginOutput, error)
	logoutFn         func(ctx context.Context, sessionID string) error
	meFn             func(ctx context.Context, userID pgtype.UUID) (*userdto.UserResponse, error)
}

func (m *mockUserSvc) Register(_ context.Context, _ userdto.RegisterRequest) (*userdto.UserResponse, error) {
	panic("not used in auth tests")
}

func (m *mockUserSvc) Login(_ context.Context, _ userdto.LoginRequest, _, _ string) (*userdto.LoginOutput, error) {
	panic("not used in auth tests")
}

func (m *mockUserSvc) LoginWithOAuth(ctx context.Context, provider, providerID, email, displayName, photoURL, ip, userAgent string) (*userdto.LoginOutput, error) {
	return m.loginWithOAuthFn(ctx, provider, providerID, email, displayName, photoURL, ip, userAgent)
}

func (m *mockUserSvc) Logout(ctx context.Context, sessionID string) error {
	return m.logoutFn(ctx, sessionID)
}

func (m *mockUserSvc) Me(ctx context.Context, userID pgtype.UUID) (*userdto.UserResponse, error) {
	return m.meFn(ctx, userID)
}

func (m *mockUserSvc) UpdateProfile(_ context.Context, _ pgtype.UUID, _ userdto.UpdateProfileRequest) (*userdto.UserResponse, error) {
	panic("not used in auth tests")
}

// testConfig returns a minimal config suitable for unit tests.
func testConfig() *config.Config {
	return &config.Config{
		GoogleClientID:     "test-google-client-id",
		GoogleClientSecret: "test-google-client-secret",
		FacebookAppID:      "test-facebook-app-id",
		FacebookAppSecret:  "test-facebook-app-secret",
		OAuthRedirectBase:  "http://localhost:8080",
		FrontendURL:        "http://localhost:3000",
	}
}

// --- OAuthStart ---

func TestOAuthStart_Google(t *testing.T) {
	svc := authsvc.New(&mockUserSvc{}, testConfig())

	url, err := svc.OAuthStart(context.Background(), "google", "test-state-abc")
	require.NoError(t, err)
	assert.Contains(t, url, "accounts.google.com")
	assert.Contains(t, url, "state=test-state-abc")
	assert.Contains(t, url, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fapi%2Fv1%2Fauth%2Fgoogle%2Fcallback")
}

func TestOAuthStart_Facebook(t *testing.T) {
	svc := authsvc.New(&mockUserSvc{}, testConfig())

	url, err := svc.OAuthStart(context.Background(), "facebook", "fb-state-xyz")
	require.NoError(t, err)
	assert.Contains(t, url, "facebook.com")
	assert.Contains(t, url, "state=fb-state-xyz")
}

func TestOAuthStart_UnknownProvider(t *testing.T) {
	svc := authsvc.New(&mockUserSvc{}, testConfig())

	_, err := svc.OAuthStart(context.Background(), "twitter", "state")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown OAuth provider")
}

func TestOAuthStart_GoogleNotConfigured(t *testing.T) {
	cfg := testConfig()
	cfg.GoogleClientID = ""
	svc := authsvc.New(&mockUserSvc{}, cfg)

	_, err := svc.OAuthStart(context.Background(), "google", "state")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestOAuthStart_FacebookNotConfigured(t *testing.T) {
	cfg := testConfig()
	cfg.FacebookAppID = ""
	svc := authsvc.New(&mockUserSvc{}, cfg)

	_, err := svc.OAuthStart(context.Background(), "facebook", "state")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// --- Me ---

func TestMe_DelegatestoUserSvc(t *testing.T) {
	wantID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	wantResp := &userdto.UserResponse{
		ID:          "00000001-0000-0000-0000-000000000000",
		Email:       "user@example.com",
		DisplayName: "Test User",
	}

	mock := &mockUserSvc{
		meFn: func(_ context.Context, userID pgtype.UUID) (*userdto.UserResponse, error) {
			assert.Equal(t, wantID, userID)
			return wantResp, nil
		},
	}

	svc := authsvc.New(mock, testConfig())
	resp, err := svc.Me(context.Background(), wantID)
	require.NoError(t, err)
	assert.Equal(t, wantResp, resp)
}

func TestMe_PropagatesError(t *testing.T) {
	wantErr := errors.New("db error")
	mock := &mockUserSvc{
		meFn: func(_ context.Context, _ pgtype.UUID) (*userdto.UserResponse, error) {
			return nil, wantErr
		},
	}

	svc := authsvc.New(mock, testConfig())
	_, err := svc.Me(context.Background(), pgtype.UUID{Valid: true})
	require.ErrorIs(t, err, wantErr)
}

// --- Logout ---

func TestLogout_DelegatesToUserSvc(t *testing.T) {
	const sessionID = "test-session-id"
	called := false
	mock := &mockUserSvc{
		logoutFn: func(_ context.Context, sid string) error {
			assert.Equal(t, sessionID, sid)
			called = true
			return nil
		},
	}

	svc := authsvc.New(mock, testConfig())
	err := svc.Logout(context.Background(), sessionID)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestLogout_PropagatesError(t *testing.T) {
	wantErr := errors.New("delete session failed")
	mock := &mockUserSvc{
		logoutFn: func(_ context.Context, _ string) error {
			return wantErr
		},
	}

	svc := authsvc.New(mock, testConfig())
	err := svc.Logout(context.Background(), "any")
	require.ErrorIs(t, err, wantErr)
}

// --- OAuthCallback (unit-testable portion: bad code triggers exchange failure) ---

func TestOAuthCallback_UnknownProvider(t *testing.T) {
	svc := authsvc.New(&mockUserSvc{}, testConfig())

	_, err := svc.OAuthCallback(context.Background(), "twitter", "code", "127.0.0.1", "go-test/1.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown OAuth provider")
}

func TestOAuthCallback_GoogleNotConfigured(t *testing.T) {
	cfg := testConfig()
	cfg.GoogleClientID = ""
	svc := authsvc.New(&mockUserSvc{}, cfg)

	_, err := svc.OAuthCallback(context.Background(), "google", "code", "127.0.0.1", "go-test/1.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestOAuthCallback_ExchangeFailureReturnsBadRequest(t *testing.T) {
	// With a real (but fake) code, Exchange will fail against Google's servers.
	// We verify that the error is wrapped as a 400 BadRequest (not a 500).
	svc := authsvc.New(&mockUserSvc{}, testConfig())

	_, err := svc.OAuthCallback(context.Background(), "google", "invalid-code", "127.0.0.1", "go-test/1.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth code exchange failed")
}

// helper: build a LoginOutput for mock responses
func makeLoginOutput() *userdto.LoginOutput {
	return &userdto.LoginOutput{
		SessionID: "sess-abc",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: &userdto.UserResponse{
			ID:          "00000001-0000-0000-0000-000000000000",
			Email:       "user@example.com",
			DisplayName: "Test User",
		},
	}
}

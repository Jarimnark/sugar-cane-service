package settingsvc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	oauthrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/oauth"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
	userrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/user"
	settingsvc "github.com/Jarimnark/sugar-cane-service/internal/service/setting"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

// ---------------------------------------------------------------------------
// Minimal in-memory fakes (implement only what the tests need)
// ---------------------------------------------------------------------------

// fakeUserRepo implements userrepo.Repository.
type fakeUserRepo struct {
	updateProfileFn func(ctx context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error)
	softDeleteFn    func(ctx context.Context, id pgtype.UUID) error
}

func (f *fakeUserRepo) Create(_ context.Context, _ sqlcgen.CreateUserParams) (sqlcgen.User, error) {
	panic("not needed in setting tests")
}
func (f *fakeUserRepo) GetByID(_ context.Context, _ pgtype.UUID) (sqlcgen.User, error) {
	panic("not needed in setting tests")
}
func (f *fakeUserRepo) GetByEmail(_ context.Context, _ string) (sqlcgen.User, error) {
	panic("not needed in setting tests")
}
func (f *fakeUserRepo) UpdateProfile(ctx context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error) {
	return f.updateProfileFn(ctx, params)
}
func (f *fakeUserRepo) VerifyEmail(_ context.Context, _ pgtype.UUID) error {
	panic("not needed in setting tests")
}
func (f *fakeUserRepo) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	return f.softDeleteFn(ctx, id)
}

var _ userrepo.Repository = (*fakeUserRepo)(nil)

// fakeOAuthRepo implements oauthrepo.Repository.
type fakeOAuthRepo struct {
	deleteByUserAndProviderFn func(ctx context.Context, userID pgtype.UUID, provider string) error
}

func (f *fakeOAuthRepo) GetByProvider(_ context.Context, _, _ string) (sqlcgen.OauthAccount, error) {
	panic("not needed in setting tests")
}
func (f *fakeOAuthRepo) Create(_ context.Context, _ pgtype.UUID, _, _ string) (sqlcgen.OauthAccount, error) {
	panic("not needed in setting tests")
}
func (f *fakeOAuthRepo) DeleteByUserAndProvider(ctx context.Context, userID pgtype.UUID, provider string) error {
	return f.deleteByUserAndProviderFn(ctx, userID, provider)
}

var _ oauthrepo.Repository = (*fakeOAuthRepo)(nil)

// fakeSessionRepo implements sessionrepo.Repository.
type fakeSessionRepo struct {
	deleteAllForUserFn func(ctx context.Context, userID pgtype.UUID) error
}

func (f *fakeSessionRepo) Create(_ context.Context, _ sqlcgen.CreateSessionParams) (sqlcgen.Session, error) {
	panic("not needed in setting tests")
}
func (f *fakeSessionRepo) Get(_ context.Context, _ string) (sqlcgen.Session, error) {
	panic("not needed in setting tests")
}
func (f *fakeSessionRepo) Touch(_ context.Context, _ string) error {
	panic("not needed in setting tests")
}
func (f *fakeSessionRepo) Delete(_ context.Context, _ string) error {
	panic("not needed in setting tests")
}
func (f *fakeSessionRepo) DeleteAllForUser(ctx context.Context, userID pgtype.UUID) error {
	return f.deleteAllForUserFn(ctx, userID)
}

var _ sessionrepo.Repository = (*fakeSessionRepo)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testUserID() pgtype.UUID {
	var id pgtype.UUID
	_ = id.Scan("11111111-1111-1111-1111-111111111111")
	return id
}

func stubUser(displayName, bio string) sqlcgen.User {
	return sqlcgen.User{
		ID:            testUserID(),
		Email:         "test@example.com",
		DisplayName:   displayName,
		Bio:           pgtype.Text{String: bio, Valid: bio != ""},
		EmailVerified: true,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

// newSvc builds a setting service with the provided fakes and nil pool.
// Tests that exercise DeleteAccount must provide their own pool — see TestDeleteAccount.
func newSvc(users userrepo.Repository, oauths oauthrepo.Repository, sessions sessionrepo.Repository) settingsvc.Service {
	return settingsvc.New(users, oauths, sessions, nil)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateProfile_DelegatesToUpdateMe(t *testing.T) {
	called := false
	users := &fakeUserRepo{
		updateProfileFn: func(_ context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error) {
			called = true
			return stubUser(params.DisplayName, params.Bio.String), nil
		},
	}

	svc := newSvc(users, nil, nil)
	resp, err := svc.CreateProfile(context.Background(), testUserID(), userdto.UpdateProfileRequest{
		DisplayName: "Alice",
		Bio:         "Hello world",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("UpdateProfile was not called")
	}
	if resp.DisplayName != "Alice" {
		t.Errorf("want DisplayName=Alice, got %q", resp.DisplayName)
	}
}

func TestUpdateMe_Success(t *testing.T) {
	users := &fakeUserRepo{
		updateProfileFn: func(_ context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error) {
			return stubUser(params.DisplayName, params.Bio.String), nil
		},
	}

	svc := newSvc(users, nil, nil)
	resp, err := svc.UpdateMe(context.Background(), testUserID(), userdto.UpdateProfileRequest{
		DisplayName: "Bob",
		Bio:         "I make games",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DisplayName != "Bob" {
		t.Errorf("want DisplayName=Bob, got %q", resp.DisplayName)
	}
	if resp.Bio != "I make games" {
		t.Errorf("want Bio='I make games', got %q", resp.Bio)
	}
}

func TestUpdateMe_EmptyBio_ValidField(t *testing.T) {
	users := &fakeUserRepo{
		updateProfileFn: func(_ context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error) {
			// Bio should NOT be marked valid when empty.
			if params.Bio.Valid {
				return sqlcgen.User{}, fmt.Errorf("Bio.Valid should be false for empty bio")
			}
			return stubUser(params.DisplayName, ""), nil
		},
	}

	svc := newSvc(users, nil, nil)
	_, err := svc.UpdateMe(context.Background(), testUserID(), userdto.UpdateProfileRequest{
		DisplayName: "Carol",
		Bio:         "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateMe_RepoError_WrapsError(t *testing.T) {
	repoErr := errors.New("db failure")
	users := &fakeUserRepo{
		updateProfileFn: func(_ context.Context, _ sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error) {
			return sqlcgen.User{}, repoErr
		},
	}

	svc := newSvc(users, nil, nil)
	_, err := svc.UpdateMe(context.Background(), testUserID(), userdto.UpdateProfileRequest{DisplayName: "Dave"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("want wrapped repoErr, got %v", err)
	}
}

func TestAddConnection_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(nil, nil, nil)
	err := svc.AddConnection(context.Background(), testUserID(), "google")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ae, ok := domain.AsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != 400 {
		t.Errorf("want HTTP 400, got %d", ae.Code)
	}
}

func TestRemoveConnection_Success(t *testing.T) {
	var capturedProvider string
	oauths := &fakeOAuthRepo{
		deleteByUserAndProviderFn: func(_ context.Context, _ pgtype.UUID, provider string) error {
			capturedProvider = provider
			return nil
		},
	}

	svc := newSvc(nil, oauths, nil)
	err := svc.RemoveConnection(context.Background(), testUserID(), "discord")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedProvider != "discord" {
		t.Errorf("want provider=discord, got %q", capturedProvider)
	}
}

func TestRemoveConnection_RepoError_WrapsError(t *testing.T) {
	repoErr := errors.New("not found")
	oauths := &fakeOAuthRepo{
		deleteByUserAndProviderFn: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return repoErr
		},
	}

	svc := newSvc(nil, oauths, nil)
	err := svc.RemoveConnection(context.Background(), testUserID(), "google")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("want wrapped repoErr, got %v", err)
	}
}

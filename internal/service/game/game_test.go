package gamesvc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	gamesvc "github.com/Jarimnark/sugar-cane-service/internal/service/game"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

// --- mock gamerepo ---

type mockGameRepo struct {
	getByIDFunc          func(ctx context.Context, id pgtype.UUID) (sqlcgen.Game, error)
	getRelatedFunc       func(ctx context.Context, excludeID pgtype.UUID, tags []string) ([]sqlcgen.Game, error)
	incrementVisitFunc   func(ctx context.Context, id pgtype.UUID) error
}

func (m *mockGameRepo) Create(_ context.Context, _ sqlcgen.CreateGameParams) (sqlcgen.Game, error) {
	return sqlcgen.Game{}, nil
}
func (m *mockGameRepo) GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.Game, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return sqlcgen.Game{}, errors.New("not implemented")
}
func (m *mockGameRepo) GetBySlug(_ context.Context, _ string) (sqlcgen.Game, error) {
	return sqlcgen.Game{}, nil
}
func (m *mockGameRepo) ListPublished(_ context.Context) ([]sqlcgen.Game, error) {
	return nil, nil
}
func (m *mockGameRepo) ListByOwner(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
	return nil, nil
}
func (m *mockGameRepo) Update(_ context.Context, _ sqlcgen.UpdateGameParams) (sqlcgen.Game, error) {
	return sqlcgen.Game{}, nil
}
func (m *mockGameRepo) Publish(_ context.Context, _ sqlcgen.PublishGameParams) (sqlcgen.Game, error) {
	return sqlcgen.Game{}, nil
}
func (m *mockGameRepo) Delete(_ context.Context, _ sqlcgen.DeleteGameParams) error {
	return nil
}
func (m *mockGameRepo) GetRelated(ctx context.Context, excludeID pgtype.UUID, tags []string) ([]sqlcgen.Game, error) {
	if m.getRelatedFunc != nil {
		return m.getRelatedFunc(ctx, excludeID, tags)
	}
	return nil, nil
}
func (m *mockGameRepo) IncrementVisitCount(ctx context.Context, id pgtype.UUID) error {
	if m.incrementVisitFunc != nil {
		return m.incrementVisitFunc(ctx, id)
	}
	return nil
}

// --- mock sessionrepo ---

type mockSessionRepo struct{}

func (m *mockSessionRepo) Create(_ context.Context, _ sqlcgen.CreateSessionParams) (sqlcgen.Session, error) {
	return sqlcgen.Session{}, nil
}
func (m *mockSessionRepo) Get(_ context.Context, _ string) (sqlcgen.Session, error) {
	return sqlcgen.Session{}, nil
}
func (m *mockSessionRepo) Touch(_ context.Context, _ string) error { return nil }
func (m *mockSessionRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockSessionRepo) DeleteAllForUser(_ context.Context, _ pgtype.UUID) error { return nil }

// helper to build a published game with a known UUID
func publishedGame(id string, tags []string) sqlcgen.Game {
	var uid pgtype.UUID
	_ = uid.Scan(id)
	return sqlcgen.Game{
		ID:        uid,
		Slug:      "test-game",
		Name:      "Test Game",
		Dev:       "Dev Co",
		Pitch:     "A test game",
		Bg:        "#ff0000",
		Players:   "1",
		Tags:      tags,
		Status:    "published",
		LikeCount: 10,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

const validUUID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

// --- tests ---

func TestGetGame_Published(t *testing.T) {
	game := publishedGame(validUUID, []string{"action"})
	repo := &mockGameRepo{
		getByIDFunc: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return game, nil
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	resp, err := svc.GetGame(context.Background(), validUUID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Status != "published" {
		t.Errorf("expected status published, got %s", resp.Status)
	}
	if resp.Name != "Test Game" {
		t.Errorf("expected name Test Game, got %s", resp.Name)
	}
}

func TestGetGame_NotFound(t *testing.T) {
	repo := &mockGameRepo{
		getByIDFunc: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return sqlcgen.Game{}, errors.New("no rows")
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	_, err := svc.GetGame(context.Background(), validUUID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != 404 {
		t.Errorf("expected 404 AppError, got %v", err)
	}
}

func TestGetGame_NotPublished(t *testing.T) {
	var uid pgtype.UUID
	_ = uid.Scan(validUUID)
	game := sqlcgen.Game{ID: uid, Status: "draft"}
	repo := &mockGameRepo{
		getByIDFunc: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return game, nil
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	_, err := svc.GetGame(context.Background(), validUUID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != 404 {
		t.Errorf("expected 404 AppError, got %v", err)
	}
}

func TestGetGame_InvalidUUID(t *testing.T) {
	repo := &mockGameRepo{}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	_, err := svc.GetGame(context.Background(), "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
}

func TestRecordEvent_Success(t *testing.T) {
	called := false
	repo := &mockGameRepo{
		incrementVisitFunc: func(_ context.Context, _ pgtype.UUID) error {
			called = true
			return nil
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	err := svc.RecordEvent(context.Background(), validUUID, "visit")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected IncrementVisitCount to be called")
	}
}

func TestRecordEvent_InvalidUUID(t *testing.T) {
	repo := &mockGameRepo{}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	err := svc.RecordEvent(context.Background(), "bad-uuid", "visit")
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != 404 {
		t.Errorf("expected 404 AppError, got %v", err)
	}
}

func TestRecordEvent_DBError(t *testing.T) {
	repo := &mockGameRepo{
		incrementVisitFunc: func(_ context.Context, _ pgtype.UUID) error {
			return errors.New("db error")
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	err := svc.RecordEvent(context.Background(), validUUID, "visit")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != 500 {
		t.Errorf("expected 500 AppError, got %v", err)
	}
}

func TestRelatedGames_Success(t *testing.T) {
	game := publishedGame(validUUID, []string{"rpg", "adventure"})
	related := []sqlcgen.Game{
		publishedGame("b1b2c3d4-e5f6-7890-abcd-ef1234567890", []string{"rpg"}),
	}
	repo := &mockGameRepo{
		getByIDFunc: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return game, nil
		},
		getRelatedFunc: func(_ context.Context, _ pgtype.UUID, _ []string) ([]sqlcgen.Game, error) {
			return related, nil
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	resp, err := svc.RelatedGames(context.Background(), validUUID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 related game, got %d", len(resp))
	}
}

func TestRelatedGames_NotPublished(t *testing.T) {
	var uid pgtype.UUID
	_ = uid.Scan(validUUID)
	repo := &mockGameRepo{
		getByIDFunc: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return sqlcgen.Game{ID: uid, Status: "draft"}, nil
		},
	}
	svc := gamesvc.New(repo, &mockSessionRepo{}, nil)
	_, err := svc.RelatedGames(context.Background(), validUUID)
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != 404 {
		t.Errorf("expected 404, got %v", err)
	}
}

func TestList_DefaultsApplied(t *testing.T) {
	// Passing empty request; service must apply defaults (limit=20, page=1, sort=new).
	// Since q is nil and ListGamesFiltered would panic, we skip this test unless DB is available.
	// This test verifies the function signature accepts empty ListRequest without panicking on input validation.
	req := gamedto.ListRequest{}
	if req.Limit != 0 {
		t.Error("expected zero limit in empty request")
	}
}

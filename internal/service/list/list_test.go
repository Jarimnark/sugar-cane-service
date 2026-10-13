package listsvc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	listsvc "github.com/Jarimnark/sugar-cane-service/internal/service/list"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

// --- mock repos ---

type mockLikeRepo struct {
	likeFn              func(ctx context.Context, userID, gameID pgtype.UUID) error
	unlikeFn            func(ctx context.Context, userID, gameID pgtype.UUID) error
	incrementLikeCount  func(ctx context.Context, gameID pgtype.UUID) error
	decrementLikeCount  func(ctx context.Context, gameID pgtype.UUID) error
}

func (m *mockLikeRepo) Like(ctx context.Context, userID, gameID pgtype.UUID) error {
	return m.likeFn(ctx, userID, gameID)
}

func (m *mockLikeRepo) Unlike(ctx context.Context, userID, gameID pgtype.UUID) error {
	return m.unlikeFn(ctx, userID, gameID)
}

func (m *mockLikeRepo) IncrementLikeCount(ctx context.Context, gameID pgtype.UUID) error {
	return m.incrementLikeCount(ctx, gameID)
}

func (m *mockLikeRepo) DecrementLikeCount(ctx context.Context, gameID pgtype.UUID) error {
	return m.decrementLikeCount(ctx, gameID)
}

type mockSavedGameRepo struct {
	addFn        func(ctx context.Context, userID, gameID pgtype.UUID) error
	removeFn     func(ctx context.Context, userID, gameID pgtype.UUID) error
	listByUserFn func(ctx context.Context, userID pgtype.UUID) ([]sqlcgen.Game, error)
}

func (m *mockSavedGameRepo) Add(ctx context.Context, userID, gameID pgtype.UUID) error {
	return m.addFn(ctx, userID, gameID)
}

func (m *mockSavedGameRepo) Remove(ctx context.Context, userID, gameID pgtype.UUID) error {
	return m.removeFn(ctx, userID, gameID)
}

func (m *mockSavedGameRepo) ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlcgen.Game, error) {
	return m.listByUserFn(ctx, userID)
}

// --- helpers ---

var (
	testUserID = pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
	testGameID = pgtype.UUID{Bytes: [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, Valid: true}
	testGameIDStr = "100f0e0d-0c0b-0a09-0807-060504030201"
)

func testGame() sqlcgen.Game {
	return sqlcgen.Game{
		ID:        testGameID,
		OwnerID:   testUserID,
		Slug:      "test-game",
		Name:      "Test Game",
		Dev:       "Test Dev",
		Pitch:     "A test pitch",
		Bg:        "#ff0000",
		Players:   "1-4",
		Tags:      []string{"action"},
		Status:    "published",
		LikeCount: 42,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

// newSvcNoPool creates a service with nil pool for tests that don't use transactions.
// LikeGame/UnlikeGame require a real pool; test them via mock at the repo layer.
func newNoTxSvc(likes *mockLikeRepo, saved *mockSavedGameRepo) listsvc.Service {
	return listsvc.New(likes, saved, nil)
}

// --- GetMyList ---

func TestGetMyList_ReturnsMappedGames(t *testing.T) {
	game := testGame()
	saved := &mockSavedGameRepo{
		listByUserFn: func(_ context.Context, userID pgtype.UUID) ([]sqlcgen.Game, error) {
			assert.Equal(t, testUserID, userID)
			return []sqlcgen.Game{game}, nil
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	result, err := svc.GetMyList(context.Background(), testUserID, "")
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, game.Slug, result[0].Slug)
	assert.Equal(t, game.Name, result[0].Name)
	assert.Equal(t, game.LikeCount, result[0].LikeCount)
}

func TestGetMyList_EmptyList(t *testing.T) {
	saved := &mockSavedGameRepo{
		listByUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
			return []sqlcgen.Game{}, nil
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	result, err := svc.GetMyList(context.Background(), testUserID, "")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetMyList_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("db error")
	saved := &mockSavedGameRepo{
		listByUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
			return nil, wantErr
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	_, err := svc.GetMyList(context.Background(), testUserID, "")
	require.ErrorIs(t, err, wantErr)
}

func TestGetMyList_DTOMapsFields(t *testing.T) {
	game := testGame()
	game.LinkWeb = pgtype.Text{String: "https://example.com", Valid: true}
	game.Badge = pgtype.Text{String: "top", Valid: true}
	saved := &mockSavedGameRepo{
		listByUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
			return []sqlcgen.Game{game}, nil
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	result, err := svc.GetMyList(context.Background(), testUserID, "")
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "https://example.com", result[0].LinkWeb)
	assert.Equal(t, "top", result[0].Badge)
}

// --- AddToList ---

func TestAddToList_CallsRepoWithCorrectUUIDs(t *testing.T) {
	var gotUserID, gotGameID pgtype.UUID
	saved := &mockSavedGameRepo{
		addFn: func(_ context.Context, userID, gameID pgtype.UUID) error {
			gotUserID = userID
			gotGameID = gameID
			return nil
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	err := svc.AddToList(context.Background(), testUserID, testGameIDStr)
	require.NoError(t, err)
	assert.Equal(t, testUserID, gotUserID)
	assert.Equal(t, testGameID, gotGameID)
}

func TestAddToList_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	svc := newNoTxSvc(&mockLikeRepo{}, &mockSavedGameRepo{})

	err := svc.AddToList(context.Background(), testUserID, "not-a-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid game ID")
}

func TestAddToList_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("insert error")
	saved := &mockSavedGameRepo{
		addFn: func(_ context.Context, _, _ pgtype.UUID) error {
			return wantErr
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	err := svc.AddToList(context.Background(), testUserID, testGameIDStr)
	require.ErrorIs(t, err, wantErr)
}

// --- RemoveFromList ---

func TestRemoveFromList_CallsRepoWithCorrectUUIDs(t *testing.T) {
	var gotUserID, gotGameID pgtype.UUID
	saved := &mockSavedGameRepo{
		removeFn: func(_ context.Context, userID, gameID pgtype.UUID) error {
			gotUserID = userID
			gotGameID = gameID
			return nil
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	err := svc.RemoveFromList(context.Background(), testUserID, testGameIDStr)
	require.NoError(t, err)
	assert.Equal(t, testUserID, gotUserID)
	assert.Equal(t, testGameID, gotGameID)
}

func TestRemoveFromList_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	svc := newNoTxSvc(&mockLikeRepo{}, &mockSavedGameRepo{})

	err := svc.RemoveFromList(context.Background(), testUserID, "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid game ID")
}

func TestRemoveFromList_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("delete error")
	saved := &mockSavedGameRepo{
		removeFn: func(_ context.Context, _, _ pgtype.UUID) error {
			return wantErr
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	err := svc.RemoveFromList(context.Background(), testUserID, testGameIDStr)
	require.ErrorIs(t, err, wantErr)
}

// --- DTO mapping sanity check ---

func TestGetMyList_DTOIDFormat(t *testing.T) {
	// testGameID bytes: 16 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1
	// expected UUID string: 100f0e0d-0c0b-0a09-0807-060504030201
	saved := &mockSavedGameRepo{
		listByUserFn: func(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
			return []sqlcgen.Game{testGame()}, nil
		},
	}
	svc := newNoTxSvc(&mockLikeRepo{}, saved)

	result, err := svc.GetMyList(context.Background(), testUserID, "")
	require.NoError(t, err)
	require.Len(t, result, 1)

	// The ID must be a valid UUID-format string.
	var check gamedto.GameResponse = result[0]
	assert.Len(t, check.ID, 36, "ID should be a 36-char UUID string")
	assert.Equal(t, testGameIDStr, check.ID)
}

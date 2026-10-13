package devsvc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devsvc "github.com/Jarimnark/sugar-cane-service/internal/service/dev"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

// --- mock game repository ---

type mockGameRepo struct {
	createFn           func(ctx context.Context, params sqlcgen.CreateGameParams) (sqlcgen.Game, error)
	getByIDFn          func(ctx context.Context, id pgtype.UUID) (sqlcgen.Game, error)
	getBySlugFn        func(ctx context.Context, slug string) (sqlcgen.Game, error)
	listPublishedFn    func(ctx context.Context) ([]sqlcgen.Game, error)
	listByOwnerFn      func(ctx context.Context, ownerID pgtype.UUID) ([]sqlcgen.Game, error)
	updateFn           func(ctx context.Context, params sqlcgen.UpdateGameParams) (sqlcgen.Game, error)
	publishFn          func(ctx context.Context, params sqlcgen.PublishGameParams) (sqlcgen.Game, error)
	deleteFn           func(ctx context.Context, params sqlcgen.DeleteGameParams) error
	getRelatedFn       func(ctx context.Context, excludeID pgtype.UUID, tags []string) ([]sqlcgen.Game, error)
	incrementVisitFn   func(ctx context.Context, id pgtype.UUID) error
}

func (m *mockGameRepo) Create(ctx context.Context, params sqlcgen.CreateGameParams) (sqlcgen.Game, error) {
	return m.createFn(ctx, params)
}

func (m *mockGameRepo) GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.Game, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockGameRepo) GetBySlug(ctx context.Context, slug string) (sqlcgen.Game, error) {
	return m.getBySlugFn(ctx, slug)
}

func (m *mockGameRepo) ListPublished(ctx context.Context) ([]sqlcgen.Game, error) {
	return m.listPublishedFn(ctx)
}

func (m *mockGameRepo) ListByOwner(ctx context.Context, ownerID pgtype.UUID) ([]sqlcgen.Game, error) {
	return m.listByOwnerFn(ctx, ownerID)
}

func (m *mockGameRepo) Update(ctx context.Context, params sqlcgen.UpdateGameParams) (sqlcgen.Game, error) {
	return m.updateFn(ctx, params)
}

func (m *mockGameRepo) Publish(ctx context.Context, params sqlcgen.PublishGameParams) (sqlcgen.Game, error) {
	return m.publishFn(ctx, params)
}

func (m *mockGameRepo) Delete(ctx context.Context, params sqlcgen.DeleteGameParams) error {
	return m.deleteFn(ctx, params)
}

func (m *mockGameRepo) GetRelated(ctx context.Context, excludeID pgtype.UUID, tags []string) ([]sqlcgen.Game, error) {
	return m.getRelatedFn(ctx, excludeID, tags)
}

func (m *mockGameRepo) IncrementVisitCount(ctx context.Context, id pgtype.UUID) error {
	return m.incrementVisitFn(ctx, id)
}

// --- helpers ---

var (
	ownerID = pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	gameID  = pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
)

func makeGame(status string) sqlcgen.Game {
	return sqlcgen.Game{
		ID:      gameID,
		OwnerID: ownerID,
		Slug:    "my-game",
		Name:    "My Game",
		Dev:     "Dev Studio",
		Pitch:   "A short pitch for the game.",
		Description: []byte("A longer description of the game."),
		Bg:      "#ffffff",
		Players: "1-4",
		Tags:    []string{"action", "rpg"},
		Status:  status,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

// --- ListMyGames ---

func TestListMyGames_ReturnsGames(t *testing.T) {
	repo := &mockGameRepo{
		listByOwnerFn: func(_ context.Context, uid pgtype.UUID) ([]sqlcgen.Game, error) {
			assert.Equal(t, ownerID, uid)
			return []sqlcgen.Game{makeGame("draft"), makeGame("published")}, nil
		},
	}
	svc := devsvc.New(repo)

	result, err := svc.ListMyGames(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, gameID.String(), result[0].ID)
}

func TestListMyGames_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &mockGameRepo{
		listByOwnerFn: func(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
			return nil, nil
		},
	}
	svc := devsvc.New(repo)

	result, err := svc.ListMyGames(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestListMyGames_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("db failure")
	repo := &mockGameRepo{
		listByOwnerFn: func(_ context.Context, _ pgtype.UUID) ([]sqlcgen.Game, error) {
			return nil, wantErr
		},
	}
	svc := devsvc.New(repo)

	_, err := svc.ListMyGames(context.Background(), ownerID)
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

// --- CreateOrUpdateGame (create path) ---

func TestCreateOrUpdateGame_Create_Success(t *testing.T) {
	createdGame := makeGame("draft")
	repo := &mockGameRepo{
		createFn: func(_ context.Context, p sqlcgen.CreateGameParams) (sqlcgen.Game, error) {
			assert.Equal(t, ownerID, p.OwnerID)
			assert.Equal(t, "my-game", p.Slug)
			assert.Equal(t, "My Game", p.Name)
			return createdGame, nil
		},
	}
	svc := devsvc.New(repo)

	req := gamedto.CreateOrUpdateGameRequest{
		Slug:        "my-game",
		Name:        "My Game",
		Dev:         "Dev Studio",
		Pitch:       "A short pitch for the game.",
		Description: "A longer description of the game.",
		Bg:          "#ffffff",
		Players:     "1-4",
		Tags:        []string{"action"},
	}
	resp, err := svc.CreateOrUpdateGame(context.Background(), ownerID, req, false)
	require.NoError(t, err)
	assert.Equal(t, gameID.String(), resp.ID)
	assert.Equal(t, "draft", resp.Status)
}

func TestCreateOrUpdateGame_Create_ThenPublish(t *testing.T) {
	createdGame := makeGame("draft")
	publishedGame := makeGame("published")
	repo := &mockGameRepo{
		createFn: func(_ context.Context, _ sqlcgen.CreateGameParams) (sqlcgen.Game, error) {
			return createdGame, nil
		},
		publishFn: func(_ context.Context, p sqlcgen.PublishGameParams) (sqlcgen.Game, error) {
			assert.Equal(t, gameID, p.ID)
			assert.Equal(t, ownerID, p.OwnerID)
			return publishedGame, nil
		},
	}
	svc := devsvc.New(repo)

	req := gamedto.CreateOrUpdateGameRequest{
		Slug:        "my-game",
		Name:        "My Game",
		Dev:         "Dev Studio",
		Pitch:       "A short pitch for the game.",
		Description: "A longer description of the game.",
		Bg:          "#ffffff",
		Players:     "1-4",
	}
	resp, err := svc.CreateOrUpdateGame(context.Background(), ownerID, req, true)
	require.NoError(t, err)
	assert.Equal(t, "published", resp.Status)
}

func TestCreateOrUpdateGame_Create_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("insert failed")
	repo := &mockGameRepo{
		createFn: func(_ context.Context, _ sqlcgen.CreateGameParams) (sqlcgen.Game, error) {
			return sqlcgen.Game{}, wantErr
		},
	}
	svc := devsvc.New(repo)

	req := gamedto.CreateOrUpdateGameRequest{
		Name:        "My Game",
		Dev:         "Dev Studio",
		Pitch:       "A short pitch for the game.",
		Description: "A longer description of the game.",
		Bg:          "#ffffff",
		Players:     "1-4",
	}
	_, err := svc.CreateOrUpdateGame(context.Background(), ownerID, req, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

// --- CreateOrUpdateGame (update path) ---

func TestCreateOrUpdateGame_Update_Success(t *testing.T) {
	updatedGame := makeGame("draft")
	repo := &mockGameRepo{
		updateFn: func(_ context.Context, p sqlcgen.UpdateGameParams) (sqlcgen.Game, error) {
			assert.Equal(t, gameID, p.ID)
			assert.Equal(t, ownerID, p.OwnerID)
			assert.Equal(t, "Updated Name", p.Name)
			return updatedGame, nil
		},
	}
	svc := devsvc.New(repo)

	req := gamedto.CreateOrUpdateGameRequest{
		ID:          gameID.String(),
		Name:        "Updated Name",
		Dev:         "Dev Studio",
		Pitch:       "A short pitch for the game.",
		Description: "A longer description of the game.",
		Bg:          "#ffffff",
		Players:     "1-4",
	}
	resp, err := svc.CreateOrUpdateGame(context.Background(), ownerID, req, false)
	require.NoError(t, err)
	assert.Equal(t, gameID.String(), resp.ID)
}

func TestCreateOrUpdateGame_Update_InvalidID(t *testing.T) {
	repo := &mockGameRepo{}
	svc := devsvc.New(repo)

	req := gamedto.CreateOrUpdateGameRequest{
		ID:          "not-a-uuid",
		Name:        "My Game",
		Dev:         "Dev Studio",
		Pitch:       "A short pitch for the game.",
		Description: "A longer description of the game.",
		Bg:          "#ffffff",
		Players:     "1-4",
	}
	_, err := svc.CreateOrUpdateGame(context.Background(), ownerID, req, false)
	require.Error(t, err)
}

func TestCreateOrUpdateGame_AlreadyPublished_SkipsPublish(t *testing.T) {
	alreadyPublished := makeGame("published")
	publishCalled := false
	repo := &mockGameRepo{
		updateFn: func(_ context.Context, _ sqlcgen.UpdateGameParams) (sqlcgen.Game, error) {
			return alreadyPublished, nil
		},
		publishFn: func(_ context.Context, _ sqlcgen.PublishGameParams) (sqlcgen.Game, error) {
			publishCalled = true
			return alreadyPublished, nil
		},
	}
	svc := devsvc.New(repo)

	req := gamedto.CreateOrUpdateGameRequest{
		ID:          gameID.String(),
		Name:        "My Game",
		Dev:         "Dev Studio",
		Pitch:       "A short pitch for the game.",
		Description: "A longer description of the game.",
		Bg:          "#ffffff",
		Players:     "1-4",
	}
	// publish=true but game is already published — Publish should not be called
	resp, err := svc.CreateOrUpdateGame(context.Background(), ownerID, req, true)
	require.NoError(t, err)
	assert.False(t, publishCalled, "Publish should not be called when game is already published")
	assert.Equal(t, "published", resp.Status)
}

// --- GetMyGame ---

func TestGetMyGame_Success(t *testing.T) {
	game := makeGame("draft")
	repo := &mockGameRepo{
		getByIDFn: func(_ context.Context, id pgtype.UUID) (sqlcgen.Game, error) {
			assert.Equal(t, gameID, id)
			return game, nil
		},
	}
	svc := devsvc.New(repo)

	resp, err := svc.GetMyGame(context.Background(), ownerID, gameID.String())
	require.NoError(t, err)
	assert.Equal(t, gameID.String(), resp.ID)
}

func TestGetMyGame_InvalidID(t *testing.T) {
	repo := &mockGameRepo{}
	svc := devsvc.New(repo)

	_, err := svc.GetMyGame(context.Background(), ownerID, "not-a-uuid")
	require.Error(t, err)
}

func TestGetMyGame_ForbiddenWhenNotOwner(t *testing.T) {
	otherOwner := pgtype.UUID{Bytes: [16]byte{99}, Valid: true}
	game := makeGame("draft") // game.OwnerID == ownerID
	repo := &mockGameRepo{
		getByIDFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return game, nil
		},
	}
	svc := devsvc.New(repo)

	_, err := svc.GetMyGame(context.Background(), otherOwner, gameID.String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "don't own")
}

func TestGetMyGame_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("not found")
	repo := &mockGameRepo{
		getByIDFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.Game, error) {
			return sqlcgen.Game{}, wantErr
		},
	}
	svc := devsvc.New(repo)

	_, err := svc.GetMyGame(context.Background(), ownerID, gameID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

// --- DeleteMyGame ---

func TestDeleteMyGame_Success(t *testing.T) {
	repo := &mockGameRepo{
		deleteFn: func(_ context.Context, p sqlcgen.DeleteGameParams) error {
			assert.Equal(t, gameID, p.ID)
			assert.Equal(t, ownerID, p.OwnerID)
			return nil
		},
	}
	svc := devsvc.New(repo)

	err := svc.DeleteMyGame(context.Background(), ownerID, gameID.String())
	require.NoError(t, err)
}

func TestDeleteMyGame_InvalidID(t *testing.T) {
	repo := &mockGameRepo{}
	svc := devsvc.New(repo)

	err := svc.DeleteMyGame(context.Background(), ownerID, "bad-uuid")
	require.Error(t, err)
}

func TestDeleteMyGame_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("delete failed")
	repo := &mockGameRepo{
		deleteFn: func(_ context.Context, _ sqlcgen.DeleteGameParams) error {
			return wantErr
		},
	}
	svc := devsvc.New(repo)

	err := svc.DeleteMyGame(context.Background(), ownerID, gameID.String())
	require.ErrorIs(t, err, wantErr)
}

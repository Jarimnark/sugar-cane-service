package commentsvc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commentsvc "github.com/Jarimnark/sugar-cane-service/internal/service/comment"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	commentdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/comment"
)

// --- mock repo ---

type mockCommentRepo struct {
	gameExistsFn    func(ctx context.Context, gameID pgtype.UUID) (bool, error)
	countFn         func(ctx context.Context, gameID pgtype.UUID) (int64, error)
	listNewestFn    func(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameNewestRow, error)
	listTopFn       func(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameTopRow, error)
	getByIDFn       func(ctx context.Context, id pgtype.UUID) (sqlcgen.GameComment, error)
	getWithAuthorFn func(ctx context.Context, id pgtype.UUID) (sqlcgen.GetCommentWithAuthorRow, error)
	insertFn        func(ctx context.Context, gameID, userID pgtype.UUID, content string, anonymous bool) (sqlcgen.GameComment, error)
	deleteFn        func(ctx context.Context, id, userID pgtype.UUID) error
}

func (m *mockCommentRepo) GameExists(ctx context.Context, gameID pgtype.UUID) (bool, error) {
	return m.gameExistsFn(ctx, gameID)
}
func (m *mockCommentRepo) Count(ctx context.Context, gameID pgtype.UUID) (int64, error) {
	return m.countFn(ctx, gameID)
}
func (m *mockCommentRepo) ListNewest(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
	return m.listNewestFn(ctx, gameID, viewerID, limit, offset)
}
func (m *mockCommentRepo) ListTop(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameTopRow, error) {
	return m.listTopFn(ctx, gameID, viewerID, limit, offset)
}
func (m *mockCommentRepo) GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.GameComment, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCommentRepo) GetWithAuthor(ctx context.Context, id pgtype.UUID) (sqlcgen.GetCommentWithAuthorRow, error) {
	return m.getWithAuthorFn(ctx, id)
}
func (m *mockCommentRepo) Insert(ctx context.Context, gameID, userID pgtype.UUID, content string, anonymous bool) (sqlcgen.GameComment, error) {
	return m.insertFn(ctx, gameID, userID, content, anonymous)
}
func (m *mockCommentRepo) Delete(ctx context.Context, id, userID pgtype.UUID) error {
	return m.deleteFn(ctx, id, userID)
}

// Like/Unlike/Increment/Decrement are only exercised in paths that require a real pool.
func (m *mockCommentRepo) Like(_ context.Context, _, _ pgtype.UUID) error {
	return nil
}
func (m *mockCommentRepo) Unlike(_ context.Context, _, _ pgtype.UUID) error {
	return nil
}
func (m *mockCommentRepo) IncrementLikeCount(_ context.Context, _ pgtype.UUID) error {
	return nil
}
func (m *mockCommentRepo) DecrementLikeCount(_ context.Context, _ pgtype.UUID) error {
	return nil
}

// --- helpers ---

var (
	testGameID   = pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
	testGameIDS  = "01020304-0506-0708-090a-0b0c0d0e0f10"
	testOwnerID  = pgtype.UUID{Bytes: [16]byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32}, Valid: true}
	testUserID   = pgtype.UUID{Bytes: [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, Valid: true}
	testUserIDS  = "100f0e0d-0c0b-0a09-0807-060504030201"
	testCommentID  = pgtype.UUID{Bytes: [16]byte{99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99}, Valid: true}
	testCommentIDS = "63636363-6363-6363-6363-636363636363"
	testCreatedAt  = pgtype.Timestamptz{Time: time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC), Valid: true}
)

func newSvc(repo *mockCommentRepo) commentsvc.Service {
	return commentsvc.New(repo, nil)
}

// --- List ---

func TestList_InvalidGameID_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(&mockCommentRepo{})
	_, err := svc.List(context.Background(), "not-a-uuid", "", 1, 20, pgtype.UUID{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestList_InvalidSort_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(&mockCommentRepo{})
	_, err := svc.List(context.Background(), testGameIDS, "garbage", 1, 20, pgtype.UUID{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort")
}

func TestList_GameNotFound_Returns404(t *testing.T) {
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := newSvc(repo)
	_, err := svc.List(context.Background(), testGameIDS, "", 1, 20, pgtype.UUID{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "game not found")
}

func TestList_Newest_MapsRows(t *testing.T) {
	rows := []sqlcgen.ListCommentsForGameNewestRow{
		{
			ID:          testCommentID,
			GameID:      testGameID,
			UserID:      testUserID,
			Content:     "hi",
			Anonymous:   false,
			LikeCount:   7,
			CreatedAt:   testCreatedAt,
			AuthorName:  "Alice",
			AuthorPhotoUrl: pgtype.Text{String: "https://x/img.png", Valid: true},
			GameOwnerID: testOwnerID,
			Liked:       true,
		},
	}
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		countFn:      func(_ context.Context, _ pgtype.UUID) (int64, error) { return 1, nil },
		listNewestFn: func(_ context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
			assert.Equal(t, testGameID, gameID)
			assert.Equal(t, testUserID, viewerID)
			assert.Equal(t, int32(20), limit)
			assert.Equal(t, int32(0), offset)
			return rows, nil
		},
	}
	svc := newSvc(repo)

	resp, err := svc.List(context.Background(), testGameIDS, "", 1, 20, testUserID)
	require.NoError(t, err)
	require.Equal(t, 1, resp.Count)
	require.Len(t, resp.Items, 1)

	item := resp.Items[0]
	assert.Equal(t, testCommentIDS, item.ID)
	assert.Equal(t, "Alice", item.AuthorName)
	assert.Equal(t, "https://x/img.png", item.AuthorPhotoURL)
	assert.Equal(t, testUserIDS, item.UserID)
	assert.True(t, item.Liked)
	assert.True(t, item.IsOwn)
	assert.False(t, item.IsDeveloper)
	assert.Equal(t, 7, item.LikeCount)
}

func TestList_Anonymous_SuppressesIdentity(t *testing.T) {
	rows := []sqlcgen.ListCommentsForGameNewestRow{
		{
			ID:          testCommentID,
			GameID:      testGameID,
			UserID:      testUserID,
			Content:     "secret",
			Anonymous:   true,
			LikeCount:   0,
			CreatedAt:   testCreatedAt,
			AuthorName:  "Alice",
			AuthorPhotoUrl: pgtype.Text{String: "https://x/img.png", Valid: true},
			GameOwnerID: testUserID, // even if the author is the owner
			Liked:       false,
		},
	}
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		countFn:      func(_ context.Context, _ pgtype.UUID) (int64, error) { return 1, nil },
		listNewestFn: func(_ context.Context, _, _ pgtype.UUID, _, _ int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
			return rows, nil
		},
	}
	svc := newSvc(repo)

	// Viewer is also the author.
	resp, err := svc.List(context.Background(), testGameIDS, commentdto.SortNewest, 1, 20, testUserID)
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)

	item := resp.Items[0]
	assert.Equal(t, "", item.UserID)
	assert.Equal(t, "Anonymous", item.AuthorName)
	assert.Equal(t, "", item.AuthorPhotoURL)
	assert.False(t, item.IsDeveloper, "anonymous author must never have is_developer=true")
	assert.True(t, item.IsOwn, "author viewing own anon comment should still see is_own=true")
	assert.True(t, item.Anonymous)
}

func TestList_DeveloperBadge_SetWhenAuthorIsOwner(t *testing.T) {
	rows := []sqlcgen.ListCommentsForGameNewestRow{
		{
			ID:          testCommentID,
			GameID:      testGameID,
			UserID:      testOwnerID, // author == owner
			Content:     "dev note",
			Anonymous:   false,
			CreatedAt:   testCreatedAt,
			AuthorName:  "Owner",
			GameOwnerID: testOwnerID,
		},
	}
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		countFn:      func(_ context.Context, _ pgtype.UUID) (int64, error) { return 1, nil },
		listNewestFn: func(_ context.Context, _, _ pgtype.UUID, _, _ int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
			return rows, nil
		},
	}
	svc := newSvc(repo)

	resp, err := svc.List(context.Background(), testGameIDS, "", 1, 20, pgtype.UUID{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.True(t, resp.Items[0].IsDeveloper)
}

func TestList_UnauthViewer_LikedIsFalse(t *testing.T) {
	rows := []sqlcgen.ListCommentsForGameNewestRow{
		{
			ID:          testCommentID,
			GameID:      testGameID,
			UserID:      testUserID,
			Content:     "hi",
			CreatedAt:   testCreatedAt,
			AuthorName:  "Alice",
			GameOwnerID: testOwnerID,
			Liked:       false,
		},
	}
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		countFn:      func(_ context.Context, _ pgtype.UUID) (int64, error) { return 1, nil },
		listNewestFn: func(_ context.Context, _, viewerID pgtype.UUID, _, _ int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
			assert.False(t, viewerID.Valid)
			return rows, nil
		},
	}
	svc := newSvc(repo)

	resp, err := svc.List(context.Background(), testGameIDS, "", 1, 20, pgtype.UUID{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.False(t, resp.Items[0].IsOwn)
	assert.False(t, resp.Items[0].Liked)
}

func TestList_Top_CallsTopQuery(t *testing.T) {
	topCalled := false
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		countFn:      func(_ context.Context, _ pgtype.UUID) (int64, error) { return 0, nil },
		listTopFn: func(_ context.Context, _, _ pgtype.UUID, _, _ int32) ([]sqlcgen.ListCommentsForGameTopRow, error) {
			topCalled = true
			return nil, nil
		},
	}
	svc := newSvc(repo)

	_, err := svc.List(context.Background(), testGameIDS, commentdto.SortTop, 1, 20, pgtype.UUID{})
	require.NoError(t, err)
	assert.True(t, topCalled)
}

func TestList_DefaultsLimitAndPage(t *testing.T) {
	var gotLimit, gotOffset int32
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		countFn:      func(_ context.Context, _ pgtype.UUID) (int64, error) { return 0, nil },
		listNewestFn: func(_ context.Context, _, _ pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
			gotLimit = limit
			gotOffset = offset
			return nil, nil
		},
	}
	svc := newSvc(repo)

	_, err := svc.List(context.Background(), testGameIDS, "", 0, 0, pgtype.UUID{})
	require.NoError(t, err)
	assert.Equal(t, int32(20), gotLimit)
	assert.Equal(t, int32(0), gotOffset)
}

// --- Create ---

func TestCreate_InvalidGameID_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(&mockCommentRepo{})
	_, err := svc.Create(context.Background(), "not-a-uuid", testUserID, commentdto.CreateCommentRequest{Content: "hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestCreate_GameNotFound_Returns404(t *testing.T) {
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return false, nil },
	}
	svc := newSvc(repo)
	_, err := svc.Create(context.Background(), testGameIDS, testUserID, commentdto.CreateCommentRequest{Content: "hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "game not found")
}

func TestCreate_ReturnsMappedResponse(t *testing.T) {
	inserted := sqlcgen.GameComment{
		ID:        testCommentID,
		GameID:    testGameID,
		UserID:    testUserID,
		Content:   "hello",
		Anonymous: false,
		CreatedAt: testCreatedAt,
	}
	author := sqlcgen.GetCommentWithAuthorRow{
		ID:          testCommentID,
		GameID:      testGameID,
		UserID:      testUserID,
		Content:     "hello",
		Anonymous:   false,
		LikeCount:   0,
		CreatedAt:   testCreatedAt,
		AuthorName:  "Alice",
		AuthorPhotoUrl: pgtype.Text{String: "https://x/me.png", Valid: true},
		GameOwnerID: testOwnerID,
	}
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		insertFn: func(_ context.Context, _, _ pgtype.UUID, _ string, _ bool) (sqlcgen.GameComment, error) {
			return inserted, nil
		},
		getWithAuthorFn: func(_ context.Context, id pgtype.UUID) (sqlcgen.GetCommentWithAuthorRow, error) {
			assert.Equal(t, testCommentID, id)
			return author, nil
		},
	}
	svc := newSvc(repo)

	resp, err := svc.Create(context.Background(), testGameIDS, testUserID, commentdto.CreateCommentRequest{Content: "hello"})
	require.NoError(t, err)
	assert.Equal(t, testCommentIDS, resp.ID)
	assert.Equal(t, "Alice", resp.AuthorName)
	assert.Equal(t, "https://x/me.png", resp.AuthorPhotoURL)
	assert.True(t, resp.IsOwn)
	assert.False(t, resp.Liked)
	assert.Equal(t, 0, resp.LikeCount)
	assert.False(t, resp.IsDeveloper)
}

func TestCreate_Anonymous_SuppressesAuthor(t *testing.T) {
	author := sqlcgen.GetCommentWithAuthorRow{
		ID:          testCommentID,
		GameID:      testGameID,
		UserID:      testUserID,
		Content:     "secret",
		Anonymous:   true,
		LikeCount:   0,
		CreatedAt:   testCreatedAt,
		AuthorName:  "Alice",
		AuthorPhotoUrl: pgtype.Text{String: "https://x/me.png", Valid: true},
		GameOwnerID: testUserID,
	}
	repo := &mockCommentRepo{
		gameExistsFn: func(_ context.Context, _ pgtype.UUID) (bool, error) { return true, nil },
		insertFn: func(_ context.Context, _, _ pgtype.UUID, _ string, _ bool) (sqlcgen.GameComment, error) {
			return sqlcgen.GameComment{ID: testCommentID}, nil
		},
		getWithAuthorFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.GetCommentWithAuthorRow, error) {
			return author, nil
		},
	}
	svc := newSvc(repo)

	resp, err := svc.Create(context.Background(), testGameIDS, testUserID, commentdto.CreateCommentRequest{Content: "secret", Anonymous: true})
	require.NoError(t, err)
	assert.Equal(t, "", resp.UserID)
	assert.Equal(t, "Anonymous", resp.AuthorName)
	assert.True(t, resp.IsOwn)
	assert.False(t, resp.IsDeveloper)
	assert.True(t, resp.Anonymous)
}

// --- Delete ---

func TestDelete_InvalidID_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(&mockCommentRepo{})
	err := svc.Delete(context.Background(), "nope", testUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestDelete_NotFound_Returns404(t *testing.T) {
	repo := &mockCommentRepo{
		getByIDFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.GameComment, error) {
			return sqlcgen.GameComment{}, pgx.ErrNoRows
		},
	}
	svc := newSvc(repo)
	err := svc.Delete(context.Background(), testCommentIDS, testUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")
}

func TestDelete_NotAuthor_Returns403(t *testing.T) {
	other := pgtype.UUID{Bytes: [16]byte{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77}, Valid: true}
	repo := &mockCommentRepo{
		getByIDFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.GameComment, error) {
			return sqlcgen.GameComment{ID: testCommentID, UserID: other}, nil
		},
	}
	svc := newSvc(repo)
	err := svc.Delete(context.Background(), testCommentIDS, testUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not comment author")
}

func TestDelete_Success(t *testing.T) {
	deleted := false
	repo := &mockCommentRepo{
		getByIDFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.GameComment, error) {
			return sqlcgen.GameComment{ID: testCommentID, UserID: testUserID}, nil
		},
		deleteFn: func(_ context.Context, id, userID pgtype.UUID) error {
			assert.Equal(t, testCommentID, id)
			assert.Equal(t, testUserID, userID)
			deleted = true
			return nil
		},
	}
	svc := newSvc(repo)
	err := svc.Delete(context.Background(), testCommentIDS, testUserID)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestDelete_PropagatesRepoError(t *testing.T) {
	wantErr := errors.New("db error")
	repo := &mockCommentRepo{
		getByIDFn: func(_ context.Context, _ pgtype.UUID) (sqlcgen.GameComment, error) {
			return sqlcgen.GameComment{ID: testCommentID, UserID: testUserID}, nil
		},
		deleteFn: func(_ context.Context, _, _ pgtype.UUID) error {
			return wantErr
		},
	}
	svc := newSvc(repo)
	err := svc.Delete(context.Background(), testCommentIDS, testUserID)
	require.ErrorIs(t, err, wantErr)
}

// --- Like / Unlike (invalid id path only — happy path requires a real pool) ---

func TestLike_InvalidID_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(&mockCommentRepo{})
	err := svc.Like(context.Background(), "not-a-uuid", testUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestUnlike_InvalidID_ReturnsBadRequest(t *testing.T) {
	svc := newSvc(&mockCommentRepo{})
	err := svc.Unlike(context.Background(), "not-a-uuid", testUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

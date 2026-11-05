package commentrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/dbutil"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type repo struct {
	q *sqlcgen.Queries
}

func New(q *sqlcgen.Queries) Repository {
	return &repo{q: q}
}

func (r *repo) queries(ctx context.Context) *sqlcgen.Queries {
	if tx, ok := dbutil.TxFromCtx(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *repo) GameExists(ctx context.Context, gameID pgtype.UUID) (bool, error) {
	return r.queries(ctx).GameExists(ctx, gameID)
}

func (r *repo) Count(ctx context.Context, gameID pgtype.UUID) (int64, error) {
	return r.queries(ctx).CountCommentsForGame(ctx, gameID)
}

func (r *repo) ListNewest(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameNewestRow, error) {
	return r.queries(ctx).ListCommentsForGameNewest(ctx, sqlcgen.ListCommentsForGameNewestParams{
		GameID:  gameID,
		Column2: viewerID,
		Column3: limit,
		Column4: offset,
	})
}

func (r *repo) ListTop(ctx context.Context, gameID, viewerID pgtype.UUID, limit, offset int32) ([]sqlcgen.ListCommentsForGameTopRow, error) {
	return r.queries(ctx).ListCommentsForGameTop(ctx, sqlcgen.ListCommentsForGameTopParams{
		GameID:  gameID,
		Column2: viewerID,
		Column3: limit,
		Column4: offset,
	})
}

func (r *repo) GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.GameComment, error) {
	return r.queries(ctx).GetCommentByID(ctx, id)
}

func (r *repo) GetWithAuthor(ctx context.Context, id pgtype.UUID) (sqlcgen.GetCommentWithAuthorRow, error) {
	return r.queries(ctx).GetCommentWithAuthor(ctx, id)
}

func (r *repo) Insert(ctx context.Context, gameID, userID pgtype.UUID, content string, anonymous bool) (sqlcgen.GameComment, error) {
	return r.queries(ctx).InsertComment(ctx, sqlcgen.InsertCommentParams{
		GameID:    gameID,
		UserID:    userID,
		Content:   content,
		Anonymous: anonymous,
	})
}

func (r *repo) Delete(ctx context.Context, id, userID pgtype.UUID) error {
	return r.queries(ctx).DeleteComment(ctx, sqlcgen.DeleteCommentParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *repo) Like(ctx context.Context, userID, commentID pgtype.UUID) error {
	return r.queries(ctx).InsertCommentLike(ctx, sqlcgen.InsertCommentLikeParams{
		UserID:    userID,
		CommentID: commentID,
	})
}

func (r *repo) Unlike(ctx context.Context, userID, commentID pgtype.UUID) error {
	return r.queries(ctx).DeleteCommentLike(ctx, sqlcgen.DeleteCommentLikeParams{
		UserID:    userID,
		CommentID: commentID,
	})
}

func (r *repo) IncrementLikeCount(ctx context.Context, commentID pgtype.UUID) error {
	return r.queries(ctx).IncrementCommentLikeCount(ctx, commentID)
}

func (r *repo) DecrementLikeCount(ctx context.Context, commentID pgtype.UUID) error {
	return r.queries(ctx).DecrementCommentLikeCount(ctx, commentID)
}

package likerepo

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

func (r *repo) Like(ctx context.Context, userID, gameID pgtype.UUID) error {
	return r.queries(ctx).InsertGameLike(ctx, sqlcgen.InsertGameLikeParams{
		UserID: userID,
		GameID: gameID,
	})
}

func (r *repo) Unlike(ctx context.Context, userID, gameID pgtype.UUID) error {
	return r.queries(ctx).DeleteGameLike(ctx, sqlcgen.DeleteGameLikeParams{
		UserID: userID,
		GameID: gameID,
	})
}

func (r *repo) IncrementLikeCount(ctx context.Context, gameID pgtype.UUID) error {
	return r.queries(ctx).IncrementLikeCount(ctx, gameID)
}

func (r *repo) DecrementLikeCount(ctx context.Context, gameID pgtype.UUID) error {
	return r.queries(ctx).DecrementLikeCount(ctx, gameID)
}

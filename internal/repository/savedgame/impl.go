package savedgamerepo

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

func (r *repo) Add(ctx context.Context, userID, gameID pgtype.UUID) error {
	return r.queries(ctx).InsertSavedGame(ctx, sqlcgen.InsertSavedGameParams{
		UserID: userID,
		GameID: gameID,
	})
}

func (r *repo) Remove(ctx context.Context, userID, gameID pgtype.UUID) error {
	return r.queries(ctx).DeleteSavedGame(ctx, sqlcgen.DeleteSavedGameParams{
		UserID: userID,
		GameID: gameID,
	})
}

func (r *repo) ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlcgen.Game, error) {
	return r.queries(ctx).ListSavedGamesByUser(ctx, userID)
}

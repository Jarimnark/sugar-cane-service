package sessionrepo

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

func (r *repo) Create(ctx context.Context, params sqlcgen.CreateSessionParams) (sqlcgen.Session, error) {
	return r.queries(ctx).CreateSession(ctx, params)
}

func (r *repo) Get(ctx context.Context, id string) (sqlcgen.Session, error) {
	return r.queries(ctx).GetSession(ctx, id)
}

func (r *repo) Touch(ctx context.Context, id string) error {
	return r.queries(ctx).TouchSession(ctx, id)
}

func (r *repo) Delete(ctx context.Context, id string) error {
	return r.queries(ctx).DeleteSession(ctx, id)
}

func (r *repo) DeleteAllForUser(ctx context.Context, userID pgtype.UUID) error {
	return r.queries(ctx).DeleteUserSessions(ctx, userID)
}

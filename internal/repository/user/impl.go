package userrepo

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

// queries returns a Queries instance scoped to the active transaction, if any.
func (r *repo) queries(ctx context.Context) *sqlcgen.Queries {
	if tx, ok := dbutil.TxFromCtx(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *repo) Create(ctx context.Context, params sqlcgen.CreateUserParams) (sqlcgen.User, error) {
	return r.queries(ctx).CreateUser(ctx, params)
}

func (r *repo) GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.User, error) {
	return r.queries(ctx).GetUserByID(ctx, id)
}

func (r *repo) GetByEmail(ctx context.Context, email string) (sqlcgen.User, error) {
	return r.queries(ctx).GetUserByEmail(ctx, email)
}

func (r *repo) UpdateProfile(ctx context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error) {
	return r.queries(ctx).UpdateUserProfile(ctx, params)
}

func (r *repo) VerifyEmail(ctx context.Context, id pgtype.UUID) error {
	return r.queries(ctx).VerifyUserEmail(ctx, id)
}

func (r *repo) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	return r.queries(ctx).SoftDeleteUser(ctx, id)
}

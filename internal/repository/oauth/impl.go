package oauthrepo

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

func (r *repo) GetByProvider(ctx context.Context, provider, providerID string) (sqlcgen.OauthAccount, error) {
	return r.queries(ctx).GetOAuthAccount(ctx, sqlcgen.GetOAuthAccountParams{
		Provider:   provider,
		ProviderID: providerID,
	})
}

func (r *repo) Create(ctx context.Context, userID pgtype.UUID, provider, providerID string) (sqlcgen.OauthAccount, error) {
	return r.queries(ctx).CreateOAuthAccount(ctx, sqlcgen.CreateOAuthAccountParams{
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
	})
}

func (r *repo) DeleteByUserAndProvider(ctx context.Context, userID pgtype.UUID, provider string) error {
	return r.queries(ctx).DeleteOAuthAccountByUserAndProvider(ctx, sqlcgen.DeleteOAuthAccountByUserAndProviderParams{
		UserID:   userID,
		Provider: provider,
	})
}

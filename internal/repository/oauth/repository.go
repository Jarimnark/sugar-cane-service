package oauthrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type Repository interface {
	GetByProvider(ctx context.Context, provider, providerID string) (sqlcgen.OauthAccount, error)
	Create(ctx context.Context, userID pgtype.UUID, provider, providerID string) (sqlcgen.OauthAccount, error)
	DeleteByUserAndProvider(ctx context.Context, userID pgtype.UUID, provider string) error
}

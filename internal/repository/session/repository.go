package sessionrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type Repository interface {
	Create(ctx context.Context, params sqlcgen.CreateSessionParams) (sqlcgen.Session, error)
	Get(ctx context.Context, id string) (sqlcgen.Session, error)
	Touch(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID pgtype.UUID) error
}

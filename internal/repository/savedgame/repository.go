package savedgamerepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type Repository interface {
	Add(ctx context.Context, userID, gameID pgtype.UUID) error
	Remove(ctx context.Context, userID, gameID pgtype.UUID) error
	ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlcgen.Game, error)
}

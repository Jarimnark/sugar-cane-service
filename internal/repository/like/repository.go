package likerepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	Like(ctx context.Context, userID, gameID pgtype.UUID) error
	Unlike(ctx context.Context, userID, gameID pgtype.UUID) error
	IncrementLikeCount(ctx context.Context, gameID pgtype.UUID) error
	DecrementLikeCount(ctx context.Context, gameID pgtype.UUID) error
}

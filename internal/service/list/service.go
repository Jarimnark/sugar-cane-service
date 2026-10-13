package listsvc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type Service interface {
	LikeGame(ctx context.Context, userID pgtype.UUID, gameID string) error
	UnlikeGame(ctx context.Context, userID pgtype.UUID, gameID string) error
	GetMyList(ctx context.Context, userID pgtype.UUID, sort string) ([]gamedto.GameResponse, error)
	AddToList(ctx context.Context, userID pgtype.UUID, gameID string) error
	RemoveFromList(ctx context.Context, userID pgtype.UUID, gameID string) error
}

package devsvc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type Service interface {
	ListMyGames(ctx context.Context, userID pgtype.UUID) ([]gamedto.GameResponse, error)
	CreateOrUpdateGame(ctx context.Context, userID pgtype.UUID, req gamedto.CreateOrUpdateGameRequest, publish bool) (*gamedto.GameResponse, error)
	GetMyGame(ctx context.Context, userID pgtype.UUID, gameID string) (*gamedto.GameResponse, error)
	DeleteMyGame(ctx context.Context, userID pgtype.UUID, gameID string) error
}

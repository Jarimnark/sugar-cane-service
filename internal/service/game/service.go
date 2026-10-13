package gamesvc

import (
	"context"

	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type Service interface {
	Home(ctx context.Context) (*gamedto.HomeResponse, error)
	MasterOption(ctx context.Context) (*gamedto.MasterOptionResponse, error)
	CategoryDetail(ctx context.Context) (*gamedto.CategoryDetailResponse, error)
	List(ctx context.Context, req gamedto.ListRequest) ([]gamedto.GameResponse, error)
	Suggest(ctx context.Context, q string) ([]string, error)
	GetGame(ctx context.Context, id string) (*gamedto.GameResponse, error)
	RelatedGames(ctx context.Context, id string) ([]gamedto.GameResponse, error)
	RecordEvent(ctx context.Context, id string, eventType string) error
}

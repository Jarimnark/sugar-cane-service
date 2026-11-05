package gamerepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type Repository interface {
	Create(ctx context.Context, params sqlcgen.CreateGameParams) (sqlcgen.Game, error)
	GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.Game, error)
	GetBySlug(ctx context.Context, slug string) (sqlcgen.Game, error)
	ListPublished(ctx context.Context) ([]sqlcgen.Game, error)
	ListByOwner(ctx context.Context, ownerID pgtype.UUID) ([]sqlcgen.Game, error)
	Update(ctx context.Context, params sqlcgen.UpdateGameParams) (sqlcgen.Game, error)
	Publish(ctx context.Context, params sqlcgen.PublishGameParams) (sqlcgen.Game, error)
	Delete(ctx context.Context, params sqlcgen.DeleteGameParams) error
	GetRelated(ctx context.Context, excludeID pgtype.UUID, tags []string) ([]sqlcgen.Game, error)
	IncrementVisitCount(ctx context.Context, id pgtype.UUID) error
	UpsertGenres(ctx context.Context, gameID pgtype.UUID, genreIDs []pgtype.UUID) error
	UpsertPlatforms(ctx context.Context, gameID pgtype.UUID, platformIDs []pgtype.UUID) error
	GetGenres(ctx context.Context, gameID pgtype.UUID) ([]sqlcgen.GetGameGenresRow, error)
	GetPlatforms(ctx context.Context, gameID pgtype.UUID) ([]sqlcgen.GetGamePlatformsRow, error)
}

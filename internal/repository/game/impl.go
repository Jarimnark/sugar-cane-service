package gamerepo

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

func (r *repo) Create(ctx context.Context, params sqlcgen.CreateGameParams) (sqlcgen.Game, error) {
	return r.queries(ctx).CreateGame(ctx, params)
}

func (r *repo) GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.Game, error) {
	return r.queries(ctx).GetGameByID(ctx, id)
}

func (r *repo) GetBySlug(ctx context.Context, slug string) (sqlcgen.Game, error) {
	return r.queries(ctx).GetGameBySlug(ctx, slug)
}

func (r *repo) ListPublished(ctx context.Context) ([]sqlcgen.Game, error) {
	return r.queries(ctx).ListPublishedGames(ctx)
}

func (r *repo) ListByOwner(ctx context.Context, ownerID pgtype.UUID) ([]sqlcgen.Game, error) {
	return r.queries(ctx).ListGamesByOwner(ctx, ownerID)
}

func (r *repo) Update(ctx context.Context, params sqlcgen.UpdateGameParams) (sqlcgen.Game, error) {
	return r.queries(ctx).UpdateGame(ctx, params)
}

func (r *repo) Publish(ctx context.Context, params sqlcgen.PublishGameParams) (sqlcgen.Game, error) {
	return r.queries(ctx).PublishGame(ctx, params)
}

func (r *repo) Delete(ctx context.Context, params sqlcgen.DeleteGameParams) error {
	return r.queries(ctx).DeleteGame(ctx, params)
}

func (r *repo) GetRelated(ctx context.Context, excludeID pgtype.UUID, tags []string) ([]sqlcgen.Game, error) {
	return r.queries(ctx).GetRelatedGames(ctx, sqlcgen.GetRelatedGamesParams{
		Column1: excludeID,
		Column2: tags,
	})
}

func (r *repo) IncrementVisitCount(ctx context.Context, id pgtype.UUID) error {
	return r.queries(ctx).IncrementVisitCount(ctx, id)
}

func (r *repo) UpsertGenres(ctx context.Context, gameID pgtype.UUID, genreIDs []pgtype.UUID) error {
	q := r.queries(ctx)
	if err := q.DeleteGameGenres(ctx, gameID); err != nil {
		return err
	}
	if len(genreIDs) == 0 {
		return nil
	}
	return q.InsertGameGenres(ctx, sqlcgen.InsertGameGenresParams{
		GameID:  gameID,
		Column2: genreIDs,
	})
}

func (r *repo) UpsertPlatforms(ctx context.Context, gameID pgtype.UUID, platformIDs []pgtype.UUID) error {
	q := r.queries(ctx)
	if err := q.DeleteGamePlatforms(ctx, gameID); err != nil {
		return err
	}
	if len(platformIDs) == 0 {
		return nil
	}
	return q.InsertGamePlatforms(ctx, sqlcgen.InsertGamePlatformsParams{
		GameID:  gameID,
		Column2: platformIDs,
	})
}

func (r *repo) GetGenres(ctx context.Context, gameID pgtype.UUID) ([]sqlcgen.GetGameGenresRow, error) {
	return r.queries(ctx).GetGameGenres(ctx, gameID)
}

func (r *repo) GetPlatforms(ctx context.Context, gameID pgtype.UUID) ([]sqlcgen.GetGamePlatformsRow, error) {
	return r.queries(ctx).GetGamePlatforms(ctx, gameID)
}

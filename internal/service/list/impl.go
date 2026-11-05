package listsvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jarimnark/sugar-cane-service/internal/dbutil"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/gamemap"
	likerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/like"
	savedgamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/savedgame"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type svc struct {
	likes likerepo.Repository
	saved savedgamerepo.Repository
	pool  *pgxpool.Pool
}

func New(likes likerepo.Repository, saved savedgamerepo.Repository, pool *pgxpool.Pool) Service {
	return &svc{likes: likes, saved: saved, pool: pool}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func (s *svc) LikeGame(ctx context.Context, userID pgtype.UUID, gameID string) error {
	gid, err := parseUUID(gameID)
	if err != nil {
		return domain.BadRequest("invalid game ID", err)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("LikeGame begin: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := dbutil.WithTx(ctx, tx)
	if err := s.likes.Like(txCtx, userID, gid); err != nil {
		return fmt.Errorf("LikeGame insert: %w", err)
	}
	if err := s.likes.IncrementLikeCount(txCtx, gid); err != nil {
		return fmt.Errorf("LikeGame increment: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *svc) UnlikeGame(ctx context.Context, userID pgtype.UUID, gameID string) error {
	gid, err := parseUUID(gameID)
	if err != nil {
		return domain.BadRequest("invalid game ID", err)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("UnlikeGame begin: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := dbutil.WithTx(ctx, tx)
	if err := s.likes.Unlike(txCtx, userID, gid); err != nil {
		return fmt.Errorf("UnlikeGame delete: %w", err)
	}
	if err := s.likes.DecrementLikeCount(txCtx, gid); err != nil {
		return fmt.Errorf("UnlikeGame decrement: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *svc) GetMyList(ctx context.Context, userID pgtype.UUID) ([]gamedto.GameResponse, error) {
	games, err := s.saved.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetMyList: %w", err)
	}
	return gamemap.ToResponses(games), nil
}

func (s *svc) AddToList(ctx context.Context, userID pgtype.UUID, gameID string) error {
	gid, err := parseUUID(gameID)
	if err != nil {
		return domain.BadRequest("invalid game ID", err)
	}
	if err := s.saved.Add(ctx, userID, gid); err != nil {
		return fmt.Errorf("AddToList: %w", err)
	}
	return nil
}

func (s *svc) RemoveFromList(ctx context.Context, userID pgtype.UUID, gameID string) error {
	gid, err := parseUUID(gameID)
	if err != nil {
		return domain.BadRequest("invalid game ID", err)
	}
	if err := s.saved.Remove(ctx, userID, gid); err != nil {
		return fmt.Errorf("RemoveFromList: %w", err)
	}
	return nil
}

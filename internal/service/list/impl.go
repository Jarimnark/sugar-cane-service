package listsvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jarimnark/sugar-cane-service/internal/dbutil"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	likerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/like"
	savedgamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/savedgame"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type svc struct {
	likes  likerepo.Repository
	saved  savedgamerepo.Repository
	pool   *pgxpool.Pool
}

func New(likes likerepo.Repository, saved savedgamerepo.Repository, pool *pgxpool.Pool) Service {
	return &svc{likes: likes, saved: saved, pool: pool}
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

func (s *svc) GetMyList(ctx context.Context, userID pgtype.UUID, _ string) ([]gamedto.GameResponse, error) {
	games, err := s.saved.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetMyList: %w", err)
	}
	result := make([]gamedto.GameResponse, len(games))
	for i, g := range games {
		result[i] = gameToDTO(g)
	}
	return result, nil
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

// --- helpers ---

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func gameToDTO(g sqlcgen.Game) gamedto.GameResponse {
	var id [16]byte = g.ID.Bytes
	var ownerID [16]byte = g.OwnerID.Bytes
	return gamedto.GameResponse{
		ID:          fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:]),
		OwnerID:     fmt.Sprintf("%x-%x-%x-%x-%x", ownerID[0:4], ownerID[4:6], ownerID[6:8], ownerID[8:10], ownerID[10:]),
		Slug:        g.Slug,
		Name:        g.Name,
		Dev:         g.Dev,
		Pitch:       g.Pitch,
		Description: string(g.Description),
		LinkWeb:     g.LinkWeb.String,
		LinkAndroid: g.LinkAndroid.String,
		LinkIos:     g.LinkIos.String,
		Bg:          g.Bg,
		Players:     g.Players,
		Tags:        g.Tags,
		Status:      g.Status,
		LikeCount:   g.LikeCount,
		VisitCount:  g.VisitCount,
		Badge:       g.Badge.String,
		CreatedAt:   g.CreatedAt.Time,
	}
}

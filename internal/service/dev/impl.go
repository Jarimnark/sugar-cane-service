package devsvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	gamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/game"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type svc struct {
	games gamerepo.Repository
}

func New(games gamerepo.Repository) Service {
	return &svc{games: games}
}

// parseUUID parses a UUID string into pgtype.UUID, returning an error if invalid.
func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

// gameToDTO converts a sqlcgen.Game to a gamedto.GameResponse.
func gameToDTO(g sqlcgen.Game) gamedto.GameResponse {
	r := gamedto.GameResponse{
		ID:          g.ID.String(),
		OwnerID:     g.OwnerID.String(),
		Slug:        g.Slug,
		Name:        g.Name,
		Dev:         g.Dev,
		Pitch:       g.Pitch,
		Description: string(g.Description),
		Bg:          g.Bg,
		Players:     g.Players,
		Tags:        g.Tags,
		Status:      g.Status,
		LikeCount:   g.LikeCount,
		VisitCount:  g.VisitCount,
		CreatedAt:   g.CreatedAt.Time,
	}
	if g.LinkWeb.Valid {
		r.LinkWeb = g.LinkWeb.String
	}
	if g.LinkAndroid.Valid {
		r.LinkAndroid = g.LinkAndroid.String
	}
	if g.LinkIos.Valid {
		r.LinkIos = g.LinkIos.String
	}
	if g.Badge.Valid {
		r.Badge = g.Badge.String
	}
	return r
}

func (s *svc) ListMyGames(ctx context.Context, userID pgtype.UUID) ([]gamedto.GameResponse, error) {
	games, err := s.games.ListByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ListMyGames: %w", err)
	}
	result := make([]gamedto.GameResponse, len(games))
	for i, g := range games {
		result[i] = gameToDTO(g)
	}
	return result, nil
}

func (s *svc) CreateOrUpdateGame(ctx context.Context, userID pgtype.UUID, req gamedto.CreateOrUpdateGameRequest, publish bool) (*gamedto.GameResponse, error) {
	var game sqlcgen.Game
	var err error

	if req.ID == "" {
		// Create new game
		game, err = s.games.Create(ctx, sqlcgen.CreateGameParams{
			OwnerID:     userID,
			Slug:        req.Slug,
			Name:        req.Name,
			Dev:         req.Dev,
			Pitch:       req.Pitch,
			Description: []byte(req.Description),
			Bg:          req.Bg,
			Players:     req.Players,
			Tags:        req.Tags,
		})
		if err != nil {
			return nil, fmt.Errorf("CreateOrUpdateGame create: %w", err)
		}
	} else {
		gameID, parseErr := parseUUID(req.ID)
		if parseErr != nil {
			return nil, domain.BadRequest("invalid game ID", parseErr)
		}
		game, err = s.games.Update(ctx, sqlcgen.UpdateGameParams{
			ID:          gameID,
			OwnerID:     userID,
			Name:        req.Name,
			Pitch:       req.Pitch,
			Description: []byte(req.Description),
			Bg:          req.Bg,
			Players:     req.Players,
			Tags:        req.Tags,
			// LinkWeb, LinkAndroid, LinkIos left as zero pgtype.Text (not valid)
		})
		if err != nil {
			return nil, fmt.Errorf("CreateOrUpdateGame update: %w", err)
		}
	}

	if publish && game.Status != "published" {
		game, err = s.games.Publish(ctx, sqlcgen.PublishGameParams{
			ID:      game.ID,
			OwnerID: userID,
		})
		if err != nil {
			return nil, fmt.Errorf("CreateOrUpdateGame publish: %w", err)
		}
	}

	dto := gameToDTO(game)
	return &dto, nil
}

func (s *svc) GetMyGame(ctx context.Context, userID pgtype.UUID, gameID string) (*gamedto.GameResponse, error) {
	gid, err := parseUUID(gameID)
	if err != nil {
		return nil, domain.BadRequest("invalid game ID", err)
	}
	game, err := s.games.GetByID(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("GetMyGame: %w", err)
	}
	if game.OwnerID != userID {
		return nil, domain.Forbidden("you don't own this game")
	}
	dto := gameToDTO(game)
	return &dto, nil
}

func (s *svc) DeleteMyGame(ctx context.Context, userID pgtype.UUID, gameID string) error {
	gid, err := parseUUID(gameID)
	if err != nil {
		return domain.BadRequest("invalid game ID", err)
	}
	return s.games.Delete(ctx, sqlcgen.DeleteGameParams{
		ID:      gid,
		OwnerID: userID,
	})
}

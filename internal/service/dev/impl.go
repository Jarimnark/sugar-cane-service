package devsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/gamemap"
	gamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/game"
	userrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/user"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type svc struct {
	games gamerepo.Repository
	users userrepo.Repository
}

func New(games gamerepo.Repository, users userrepo.Repository) Service {
	return &svc{games: games, users: users}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func optionalText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func coalesceStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func (s *svc) ListMyGames(ctx context.Context, userID pgtype.UUID) ([]gamedto.GameResponse, error) {
	games, err := s.games.ListByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ListMyGames: %w", err)
	}
	return gamemap.ToResponses(games), nil
}

// resolveDev returns the dev name to persist. If the client omits it, fall back
// to the signed-in user's display_name per the new design.
func (s *svc) resolveDev(ctx context.Context, userID pgtype.UUID, reqDev string) (string, error) {
	if reqDev != "" {
		return reqDev, nil
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("resolveDev: %w", err)
	}
	return user.DisplayName, nil
}

// validatePublishReady enforces the design's publish rules. Drafts save with
// missing fields; publish requires banner, play_style, ≥1 link, ≥1 genre, ≥1
// platform.
func validatePublishReady(req gamedto.CreateOrUpdateGameRequest) error {
	var missing []string
	if req.Banner == "" {
		missing = append(missing, "banner")
	}
	if req.PlayStyle == "" {
		missing = append(missing, "play_style")
	}
	if req.LinkWeb == "" && req.LinkAndroid == "" && req.LinkIos == "" {
		missing = append(missing, "link")
	}
	if len(req.GenreIDs) == 0 {
		missing = append(missing, "genre")
	}
	if len(req.PlatformIDs) == 0 {
		missing = append(missing, "platform")
	}
	if len(missing) == 0 {
		return nil
	}
	return domain.BadRequest(fmt.Sprintf("cannot publish; missing: %v", missing), nil)
}

func (s *svc) CreateOrUpdateGame(ctx context.Context, userID pgtype.UUID, req gamedto.CreateOrUpdateGameRequest, publish bool) (*gamedto.GameResponse, error) {
	if publish {
		if err := validatePublishReady(req); err != nil {
			return nil, err
		}
	}

	dev, err := s.resolveDev(ctx, userID, req.Dev)
	if err != nil {
		return nil, err
	}

	descJSON, err := json.Marshal(req.Description)
	if err != nil {
		return nil, domain.BadRequest("invalid description", err)
	}

	var game sqlcgen.Game

	if req.ID == "" {
		if req.Slug == "" {
			return nil, domain.BadRequest("slug is required for new games", nil)
		}
		game, err = s.games.Create(ctx, sqlcgen.CreateGameParams{
			OwnerID:     userID,
			Slug:        req.Slug,
			Name:        req.Name,
			Dev:         optionalText(dev),
			Pitch:       req.Pitch,
			Description: descJSON,
			Bg:          optionalText(req.Bg),
			Players:     optionalText(req.Players),
			Tags:        coalesceStrings(req.Tags),
			VideoLinks:  coalesceStrings(req.VideoLinks),
			Screenshots: coalesceStrings(req.Screenshots),
			BannerUrl:   optionalText(req.Banner),
			PlayStyle:   optionalText(req.PlayStyle),
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
			Description: descJSON,
			LinkWeb:     optionalText(req.LinkWeb),
			LinkAndroid: optionalText(req.LinkAndroid),
			LinkIos:     optionalText(req.LinkIos),
			Bg:          optionalText(req.Bg),
			Players:     optionalText(req.Players),
			Tags:        coalesceStrings(req.Tags),
			VideoLinks:  coalesceStrings(req.VideoLinks),
			Screenshots: coalesceStrings(req.Screenshots),
			BannerUrl:   optionalText(req.Banner),
			PlayStyle:   optionalText(req.PlayStyle),
			Dev:         optionalText(dev),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.NotFound("game not found")
			}
			return nil, fmt.Errorf("CreateOrUpdateGame update: %w", err)
		}
	}

	if len(req.GenreIDs) > 0 {
		genreUUIDs := make([]pgtype.UUID, len(req.GenreIDs))
		for i, id := range req.GenreIDs {
			if err := genreUUIDs[i].Scan(id); err != nil {
				return nil, domain.BadRequest("invalid genre ID", err)
			}
		}
		if err := s.games.UpsertGenres(ctx, game.ID, genreUUIDs); err != nil {
			return nil, fmt.Errorf("UpsertGenres: %w", err)
		}
	}
	if len(req.PlatformIDs) > 0 {
		platformUUIDs := make([]pgtype.UUID, len(req.PlatformIDs))
		for i, id := range req.PlatformIDs {
			if err := platformUUIDs[i].Scan(id); err != nil {
				return nil, domain.BadRequest("invalid platform ID", err)
			}
		}
		if err := s.games.UpsertPlatforms(ctx, game.ID, platformUUIDs); err != nil {
			return nil, fmt.Errorf("UpsertPlatforms: %w", err)
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

	resp := gamemap.ToResponse(game)
	return &resp, nil
}

func (s *svc) GetMyGame(ctx context.Context, userID pgtype.UUID, gameID string) (*gamedto.GameResponse, error) {
	gid, err := parseUUID(gameID)
	if err != nil {
		return nil, domain.BadRequest("invalid game ID", err)
	}
	game, err := s.games.GetByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("game not found")
		}
		return nil, fmt.Errorf("GetMyGame: %w", err)
	}
	if game.OwnerID != userID {
		return nil, domain.Forbidden("you don't own this game")
	}
	resp := gamemap.ToResponse(game)
	return &resp, nil
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

package gamesvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/gamemap"
	gamerepo "github.com/Jarimnark/sugar-cane-service/internal/repository/game"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type svc struct {
	games    gamerepo.Repository
	sessions sessionrepo.Repository
	q        *sqlcgen.Queries
}

func New(games gamerepo.Repository, sessions sessionrepo.Repository, q *sqlcgen.Queries) Service {
	return &svc{games: games, sessions: sessions, q: q}
}

// parseOptionalUUID parses an optional UUID filter. Empty string → invalid pgtype.UUID
// (meaning "no filter"). A non-empty but malformed UUID returns a BadRequest.
func parseOptionalUUID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, domain.BadRequest("invalid UUID", err)
	}
	return u, nil
}

// parseRequiredUUID parses a required UUID path param. Any failure → NotFound,
// since we don't want to leak whether a resource exists for an unparseable ID.
func parseRequiredUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if s == "" {
		return pgtype.UUID{}, domain.NotFound("game not found")
	}
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, domain.NotFound("game not found")
	}
	return u, nil
}

func (s *svc) Home(ctx context.Context) (*gamedto.HomeResponse, error) {
	trending, err := s.q.GetHomeTrending(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("Home trending: %w", err))
	}
	popular, err := s.q.GetHomePopular(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("Home popular: %w", err))
	}
	newGames, err := s.q.GetHomeNew(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("Home new: %w", err))
	}
	return &gamedto.HomeResponse{
		Trending: gamemap.ToResponses(trending),
		Popular:  gamemap.ToResponses(popular),
		New:      gamemap.ToResponses(newGames),
	}, nil
}

func (s *svc) MasterOption(ctx context.Context) (*gamedto.MasterOptionResponse, error) {
	genres, err := s.q.ListGenres(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("MasterOption genres: %w", err))
	}
	platforms, err := s.q.ListPlatforms(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("MasterOption platforms: %w", err))
	}
	playStyles, err := s.q.ListPlayStyles(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("MasterOption play_styles: %w", err))
	}
	playerCounts, err := s.q.ListPlayerCounts(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("MasterOption player_counts: %w", err))
	}

	genreItems := make([]gamedto.OptionItem, 0, len(genres))
	for _, g := range genres {
		genreItems = append(genreItems, gamedto.OptionItem{
			ID:   g.ID.String(),
			Name: g.Name,
			Slug: g.Slug,
		})
	}

	platformItems := make([]gamedto.OptionItem, 0, len(platforms))
	for _, p := range platforms {
		platformItems = append(platformItems, gamedto.OptionItem{
			ID:   p.ID.String(),
			Name: p.Name,
			Slug: p.Slug,
		})
	}

	playStyleItems := make([]gamedto.OptionItem, 0, len(playStyles))
	for _, ps := range playStyles {
		playStyleItems = append(playStyleItems, gamedto.OptionItem{
			ID:   ps.ID.String(),
			Name: ps.Name,
			Slug: ps.Slug,
		})
	}

	playerCountItems := make([]gamedto.OptionItem, 0, len(playerCounts))
	for _, pc := range playerCounts {
		playerCountItems = append(playerCountItems, gamedto.OptionItem{
			ID:   pc.ID.String(),
			Name: pc.Name,
			Slug: pc.Slug,
		})
	}

	return &gamedto.MasterOptionResponse{
		Genres:       genreItems,
		Platforms:    platformItems,
		PlayStyles:   playStyleItems,
		PlayerCounts: playerCountItems,
	}, nil
}

func (s *svc) CategoryDetail(ctx context.Context) (*gamedto.CategoryDetailResponse, error) {
	genres, err := s.q.CountGamesByGenre(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("CategoryDetail genres: %w", err))
	}
	platforms, err := s.q.CountGamesByPlatform(ctx)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("CategoryDetail platforms: %w", err))
	}

	genreItems := make([]gamedto.CategoryItem, 0, len(genres))
	for _, g := range genres {
		genreItems = append(genreItems, gamedto.CategoryItem{
			ID:        g.ID.String(),
			Name:      g.Name,
			Slug:      g.Slug,
			GameCount: int(g.GameCount),
		})
	}

	platformItems := make([]gamedto.CategoryItem, 0, len(platforms))
	for _, p := range platforms {
		platformItems = append(platformItems, gamedto.CategoryItem{
			ID:        p.ID.String(),
			Name:      p.Name,
			Slug:      p.Slug,
			GameCount: int(p.GameCount),
		})
	}

	return &gamedto.CategoryDetailResponse{
		Genres:    genreItems,
		Platforms: platformItems,
	}, nil
}

func (s *svc) List(ctx context.Context, req gamedto.ListRequest) ([]gamedto.GameResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	sort := req.Sort
	switch sort {
	case "popular", "most_liked", "newest":
	default:
		sort = "newest"
	}

	genreID, err := parseOptionalUUID(req.GenreID)
	if err != nil {
		return nil, err
	}
	platformID, err := parseOptionalUUID(req.PlatformID)
	if err != nil {
		return nil, err
	}

	games, err := s.q.ListGamesFiltered(ctx, sqlcgen.ListGamesFilteredParams{
		Column1: genreID,
		Column2: platformID,
		Column3: sort,
		Column4: int32(limit),
		Column5: int32(offset),
		Column6: req.Q,
	})
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("List: %w", err))
	}
	return gamemap.ToResponses(games), nil
}

func (s *svc) Suggest(ctx context.Context, q string) ([]string, error) {
	rows, err := s.q.SuggestGames(ctx, q)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("Suggest: %w", err))
	}
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		names = append(names, r.Name)
	}
	return names, nil
}

func (s *svc) GetGame(ctx context.Context, id string) (*gamedto.GameResponse, error) {
	uid, err := parseRequiredUUID(id)
	if err != nil {
		return nil, err
	}
	game, err := s.games.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("game not found")
		}
		return nil, domain.Internal(fmt.Errorf("GetGame: %w", err))
	}
	if game.Status != "published" {
		return nil, domain.NotFound("game not found")
	}
	resp := gamemap.ToResponse(game)
	return &resp, nil
}

func (s *svc) RelatedGames(ctx context.Context, id string) ([]gamedto.GameResponse, error) {
	uid, err := parseRequiredUUID(id)
	if err != nil {
		return nil, err
	}
	game, err := s.games.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("game not found")
		}
		return nil, domain.Internal(fmt.Errorf("RelatedGames: %w", err))
	}
	if game.Status != "published" {
		return nil, domain.NotFound("game not found")
	}
	related, err := s.games.GetRelated(ctx, uid, game.Tags)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("RelatedGames: %w", err))
	}
	return gamemap.ToResponses(related), nil
}

func (s *svc) RecordEvent(ctx context.Context, id string, _ string) error {
	uid, err := parseRequiredUUID(id)
	if err != nil {
		return err
	}
	if err := s.games.IncrementVisitCount(ctx, uid); err != nil {
		return domain.Internal(fmt.Errorf("RecordEvent: %w", err))
	}
	return nil
}

package gamesvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
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

// gameToDTO converts a sqlcgen.Game to a gamedto.GameResponse.
func gameToDTO(g sqlcgen.Game) gamedto.GameResponse {
	r := gamedto.GameResponse{
		ID:         g.ID.String(),
		OwnerID:    g.OwnerID.String(),
		Slug:       g.Slug,
		Name:       g.Name,
		Dev:        g.Dev,
		Pitch:      g.Pitch,
		Description: string(g.Description),
		Bg:         g.Bg,
		Players:    g.Players,
		Tags:       g.Tags,
		Status:     g.Status,
		LikeCount:  g.LikeCount,
		VisitCount: g.VisitCount,
		CreatedAt:  g.CreatedAt.Time,
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

// mapGames converts a slice of sqlcgen.Game to a slice of gamedto.GameResponse.
func mapGames(games []sqlcgen.Game) []gamedto.GameResponse {
	out := make([]gamedto.GameResponse, 0, len(games))
	for _, g := range games {
		out = append(out, gameToDTO(g))
	}
	return out
}

// parsePgtypeUUID parses an optional UUID string into pgtype.UUID.
// Returns an invalid (zero-value) pgtype.UUID if s is empty.
func parsePgtypeUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	var u pgtype.UUID
	_ = u.Scan(s)
	return u
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
		Trending: mapGames(trending),
		Popular:  mapGames(popular),
		New:      mapGames(newGames),
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

	return &gamedto.MasterOptionResponse{
		Genres:    genreItems,
		Platforms: platformItems,
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
	if sort == "" {
		sort = "new"
	}

	games, err := s.q.ListGamesFiltered(ctx, sqlcgen.ListGamesFilteredParams{
		Column1: parsePgtypeUUID(req.GenreID),
		Column2: parsePgtypeUUID(req.PlatformID),
		Column3: sort,
		Column4: int32(limit),
		Column5: int32(offset),
	})
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("List: %w", err))
	}
	return mapGames(games), nil
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
	uid := parsePgtypeUUID(id)
	if !uid.Valid {
		return nil, domain.NotFound("game not found")
	}
	game, err := s.games.GetByID(ctx, uid)
	if err != nil {
		return nil, domain.NotFound("game not found")
	}
	if game.Status != "published" {
		return nil, domain.NotFound("game not found")
	}
	dto := gameToDTO(game)
	return &dto, nil
}

func (s *svc) RelatedGames(ctx context.Context, id string) ([]gamedto.GameResponse, error) {
	uid := parsePgtypeUUID(id)
	if !uid.Valid {
		return nil, domain.NotFound("game not found")
	}
	game, err := s.games.GetByID(ctx, uid)
	if err != nil {
		return nil, domain.NotFound("game not found")
	}
	if game.Status != "published" {
		return nil, domain.NotFound("game not found")
	}
	related, err := s.games.GetRelated(ctx, uid, game.Tags)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("RelatedGames: %w", err))
	}
	return mapGames(related), nil
}

func (s *svc) RecordEvent(ctx context.Context, id string, eventType string) error {
	uid := parsePgtypeUUID(id)
	if !uid.Valid {
		return domain.NotFound("game not found")
	}
	if err := s.games.IncrementVisitCount(ctx, uid); err != nil {
		return domain.Internal(fmt.Errorf("RecordEvent: %w", err))
	}
	return nil
}

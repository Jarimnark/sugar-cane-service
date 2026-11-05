package gamemap

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

var playStyleLabels = map[string]string{
	"casual":       "Casual",
	"co-op":        "Co-op",
	"pvp":          "PvP",
	"pve":          "PvE",
	"story-driven": "Story-Driven",
	"fast-reflex":  "Fast Reflex",
	"brain":        "Brain",
}

func descriptionText(raw []byte) string {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return string(raw)
	}
	return s
}

func textOrEmpty(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func ToResponse(g sqlcgen.Game) gamedto.GameResponse {
	r := gamedto.GameResponse{
		ID:          g.ID.String(),
		OwnerID:     g.OwnerID.String(),
		Slug:        g.Slug,
		Name:        g.Name,
		Dev:         textOrEmpty(g.Dev),
		Pitch:       g.Pitch,
		Description: descriptionText(g.Description),
		Bg:          textOrEmpty(g.Bg),
		Banner:      textOrEmpty(g.BannerUrl),
		PlayStyle:   textOrEmpty(g.PlayStyle),
		Players:     textOrEmpty(g.Players),
		Tags:        g.Tags,
		Status:      g.Status,
		LikeCount:   g.LikeCount,
		VisitCount:  g.VisitCount,
		CreatedAt:   g.CreatedAt.Time,
		VideoLinks:  g.VideoLinks,
		Screenshots: g.Screenshots,
	}
	if g.PlayStyle.Valid {
		r.PlayStyleLabel = playStyleLabels[g.PlayStyle.String]
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

func ToResponseWithDetails(g sqlcgen.Game, genres []sqlcgen.GetGameGenresRow, platforms []sqlcgen.GetGamePlatformsRow) gamedto.GameResponse {
	r := ToResponse(g)
	for _, genre := range genres {
		r.Genres = append(r.Genres, gamedto.OptionItem{ID: genre.ID.String(), Name: genre.Name, Slug: genre.Slug})
	}
	for _, platform := range platforms {
		r.Platforms = append(r.Platforms, gamedto.OptionItem{ID: platform.ID.String(), Name: platform.Name, Slug: platform.Slug})
	}
	return r
}

func ToResponses(games []sqlcgen.Game) []gamedto.GameResponse {
	out := make([]gamedto.GameResponse, 0, len(games))
	for _, g := range games {
		out = append(out, ToResponse(g))
	}
	return out
}

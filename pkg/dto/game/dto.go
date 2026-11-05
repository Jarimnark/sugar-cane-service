package gamedto

import "time"

type CreateGameRequest struct {
	Slug        string   `json:"slug"        validate:"required,min=2,max=255,alphanum"`
	Name        string   `json:"name"        validate:"required,min=2,max=255"`
	Dev         string   `json:"dev"         validate:"required,min=2,max=255"`
	Pitch       string   `json:"pitch"       validate:"required,min=10"`
	Description string   `json:"description" validate:"required,min=10"`
	Bg          string   `json:"bg"          validate:"required,len=7"`
	Players     string   `json:"players"     validate:"required,max=20"`
	Tags        []string `json:"tags"`
}

type UpdateGameRequest struct {
	Name        string   `json:"name"         validate:"required,min=2,max=255"`
	Pitch       string   `json:"pitch"        validate:"required,min=10"`
	Description string   `json:"description"  validate:"required,min=10"`
	LinkWeb     string   `json:"link_web"     validate:"omitempty,url"`
	LinkAndroid string   `json:"link_android" validate:"omitempty,url"`
	LinkIos     string   `json:"link_ios"     validate:"omitempty,url"`
	Bg          string   `json:"bg"           validate:"required,len=7"`
	Players     string   `json:"players"      validate:"required,max=20"`
	Tags        []string `json:"tags"`
}

// CreateOrUpdateGameRequest is used for POST /me/game.
// If ID is empty, a new game is created. If ID is set, the existing game is updated.
// Dev and Bg are now optional — the server fills Dev from the session user's
// display_name when empty, and Bg is no longer user-facing (reserved for seeds).
// Banner and PlayStyle are required at publish time; drafts may omit them.
type CreateOrUpdateGameRequest struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"         validate:"omitempty,min=2,max=255"`
	Name        string   `json:"name"         validate:"required,min=2,max=255"`
	Dev         string   `json:"dev"          validate:"omitempty,min=2,max=255"`
	Pitch       string   `json:"pitch"        validate:"required,min=10"`
	Description string   `json:"description"  validate:"required,min=10"`
	LinkWeb     string   `json:"link_web"     validate:"omitempty,url"`
	LinkAndroid string   `json:"link_android" validate:"omitempty,url"`
	LinkIos     string   `json:"link_ios"     validate:"omitempty,url"`
	Bg          string   `json:"bg"           validate:"omitempty,len=7"`
	Banner      string   `json:"banner"       validate:"omitempty,url"`
	PlayStyle   string   `json:"play_style"   validate:"omitempty,oneof=casual co-op pvp pve story-driven fast-reflex brain"`
	Players     string   `json:"players"      validate:"omitempty,oneof=single-player local-2p online-multiplayer"`
	Tags        []string `json:"tags"`
	GenreIDs    []string `json:"genre_ids"`
	PlatformIDs []string `json:"platform_ids"`
	VideoLinks  []string `json:"video_links"`
	Screenshots []string `json:"screenshots"`
}

type GameResponse struct {
	ID             string       `json:"id"`
	OwnerID        string       `json:"owner_id"`
	Slug           string       `json:"slug"`
	Name           string       `json:"name"`
	Dev            string       `json:"dev,omitempty"`
	Pitch          string       `json:"pitch"`
	Description    string       `json:"description"`
	LinkWeb        string       `json:"link_web,omitempty"`
	LinkAndroid    string       `json:"link_android,omitempty"`
	LinkIos        string       `json:"link_ios,omitempty"`
	Bg             string       `json:"bg,omitempty"`
	Banner         string       `json:"banner,omitempty"`
	PlayStyle      string       `json:"play_style,omitempty"`
	PlayStyleLabel string       `json:"play_style_label,omitempty"`
	Players        string       `json:"players"`
	Tags           []string     `json:"tags"`
	Status         string       `json:"status"`
	LikeCount      int32        `json:"like_count"`
	VisitCount     int32        `json:"visit_count"`
	Badge          string       `json:"badge,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	Genres         []OptionItem `json:"genres,omitempty"`
	Platforms      []OptionItem `json:"platforms,omitempty"`
	VideoLinks     []string     `json:"video_links,omitempty"`
	Screenshots    []string     `json:"screenshots,omitempty"`
}

type ListRequest struct {
	Sort       string `json:"sort"`
	GenreID    string `json:"genre_id"`
	PlatformID string `json:"platform_id"`
	Q          string `json:"q"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type HomeResponse struct {
	Trending []GameResponse `json:"trending"`
	Popular  []GameResponse `json:"popular"`
	New      []GameResponse `json:"new"`
}

type MasterOptionResponse struct {
	Genres       []OptionItem `json:"genres"`
	Platforms    []OptionItem `json:"platforms"`
	PlayStyles   []OptionItem `json:"play_styles"`
	PlayerCounts []OptionItem `json:"player_counts"`
}

type CategoryDetailResponse struct {
	Genres    []CategoryItem `json:"genres"`
	Platforms []CategoryItem `json:"platforms"`
}

type OptionItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CategoryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	GameCount int    `json:"game_count"`
}

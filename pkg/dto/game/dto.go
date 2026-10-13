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
type CreateOrUpdateGameRequest struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"        validate:"required_if=ID ,min=2,max=255,alphanum"`
	Name        string   `json:"name"        validate:"required,min=2,max=255"`
	Dev         string   `json:"dev"         validate:"required,min=2,max=255"`
	Pitch       string   `json:"pitch"       validate:"required,min=10"`
	Description string   `json:"description" validate:"required,min=10"`
	Bg          string   `json:"bg"          validate:"required,len=7"`
	Players     string   `json:"players"     validate:"required,max=20"`
	Tags        []string `json:"tags"`
}

type GameResponse struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Dev         string    `json:"dev"`
	Pitch       string    `json:"pitch"`
	Description string    `json:"description"`
	LinkWeb     string    `json:"link_web,omitempty"`
	LinkAndroid string    `json:"link_android,omitempty"`
	LinkIos     string    `json:"link_ios,omitempty"`
	Bg          string    `json:"bg"`
	Players     string    `json:"players"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	LikeCount   int32     `json:"like_count"`
	VisitCount  int32     `json:"visit_count"`
	Badge       string    `json:"badge,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListRequest struct {
	Sort       string `json:"sort"`
	GenreID    string `json:"genre_id"`
	PlatformID string `json:"platform_id"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type HomeResponse struct {
	Trending []GameResponse `json:"trending"`
	Popular  []GameResponse `json:"popular"`
	New      []GameResponse `json:"new"`
}

type MasterOptionResponse struct {
	Genres    []OptionItem `json:"genres"`
	Platforms []OptionItem `json:"platforms"`
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

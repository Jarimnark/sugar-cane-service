// Package hopbu defines the Hopbu game-shelf tools served over MCP.
//
// It depends only on small interfaces so it can be unit-tested with fakes;
// cmd/api wires it to the real dev/game services and S3 storage.
package hopbu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/Jarimnark/sugar-cane-service/internal/mcp"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

// Games is what the tools need from the game services, already bound to the
// MCP owner account (the user ID never comes from the MCP client).
type Games interface {
	Options(ctx context.Context) (*gamedto.MasterOptionResponse, error)
	ListMine(ctx context.Context) ([]gamedto.GameResponse, error)
	GetMine(ctx context.Context, id string) (*gamedto.GameResponse, error)
	Save(ctx context.Context, req gamedto.CreateOrUpdateGameRequest, publish bool) (*gamedto.GameResponse, error)
}

// Media stores an uploaded image and returns its public URL.
// storage.Client satisfies it.
type Media interface {
	PutObject(ctx context.Context, prefix string, reader io.Reader, size int64, contentType string) (string, error)
}

// Fetcher downloads an image from a URL.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) ([]byte, string, error)
}

// Deps are the tool dependencies. Media may be nil when storage is down;
// the upload tool then returns an error.
type Deps struct {
	Games    Games
	Media    Media
	Fetcher  Fetcher
	Validate func(any) error
}

// Instructions is sent to the client on initialize.
const Instructions = `Hopbu is a shelf of web games. These tools manage games owned by the Hopbu account this server is configured for.
Typical flow for adding a game: upload_image_from_url for the 16:9 banner (and any screenshots), then save_game with the returned URLs. Leave publish=false to save a draft the owner can review; publish=true needs banner, play_style, at least one link, a genre and a platform.
Genres and platforms can be given by name (e.g. "MMORPG", "Web game"); call list_options to see the valid values.
To edit, call save_game with the game's id and only the fields to change.`

// Tools returns the MCP tools.
func Tools(d Deps) []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_options",
			Description: "List the valid genres, platforms, play styles and player counts (with ids) for save_game.",
			InputSchema: objectSchema(nil),
			Handler: func(ctx context.Context, _ json.RawMessage) (any, error) {
				return d.Games.Options(ctx)
			},
		},
		{
			Name:        "list_my_games",
			Description: "List the account's games (drafts and published) in a compact form: id, name, status, pitch, banner, link.",
			InputSchema: objectSchema(nil),
			Handler: func(ctx context.Context, _ json.RawMessage) (any, error) {
				games, err := d.Games.ListMine(ctx)
				if err != nil {
					return nil, err
				}
				out := make([]gameSummary, 0, len(games))
				for _, g := range games {
					out = append(out, summarize(g))
				}
				return out, nil
			},
		},
		{
			Name:        "get_my_game",
			Description: "Get every field of one of the account's games.",
			InputSchema: objectSchema(map[string]any{
				"id": prop("string", "Game id from list_my_games."),
			}, "id"),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				var a struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(args, &a); err != nil || a.ID == "" {
					return nil, errors.New("id is required")
				}
				return d.Games.GetMine(ctx, a.ID)
			},
		},
		{
			Name: "save_game",
			Description: "Create a game (no id) or update one (with id). On update, only the fields you send change. " +
				"Saves a draft unless publish=true. Returns the saved game.",
			InputSchema: saveGameSchema(),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				return saveGame(ctx, d, args)
			},
		},
		{
			Name: "upload_image_from_url",
			Description: "Download an image (JPEG, PNG or WebP, max 2 MB) from a public URL and store it on Hopbu. " +
				"Returns the Hopbu URL to use as save_game banner or screenshots. Banners should be 16:9.",
			InputSchema: objectSchema(map[string]any{
				"image_url": prop("string", "Public http(s) URL of the image."),
				"kind": map[string]any{
					"type": "string", "enum": []string{"banner", "screenshot"},
					"description": "banner (default) or screenshot.",
				},
			}, "image_url"),
			Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
				return uploadFromURL(ctx, d, args)
			},
		},
	}
}

// saveArgs is CreateOrUpdateGameRequest plus MCP conveniences.
type saveArgs struct {
	gamedto.CreateOrUpdateGameRequest
	Publish   bool     `json:"publish"`
	Genres    []string `json:"genres"`
	Platforms []string `json:"platforms"`
}

func saveGame(ctx context.Context, d Deps, raw json.RawMessage) (any, error) {
	// Peek at the id first: for updates we start from the stored game and
	// lay the sent fields over it, so the caller can send only what changes.
	var peek struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &peek); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	var a saveArgs
	if peek.ID != "" {
		existing, err := d.Games.GetMine(ctx, peek.ID)
		if err != nil {
			return nil, err
		}
		a.CreateOrUpdateGameRequest = toRequest(*existing)
	}
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	req := a.CreateOrUpdateGameRequest
	req.ID = peek.ID

	if req.ID == "" && req.Slug == "" {
		req.Slug = Slugify(req.Name)
	}

	if len(a.Genres) > 0 || len(a.Platforms) > 0 {
		opts, err := d.Games.Options(ctx)
		if err != nil {
			return nil, fmt.Errorf("load options: %w", err)
		}
		if len(a.Genres) > 0 {
			ids, err := resolveOptions(a.Genres, opts.Genres, "genre")
			if err != nil {
				return nil, err
			}
			req.GenreIDs = ids
		}
		if len(a.Platforms) > 0 {
			ids, err := resolveOptions(a.Platforms, opts.Platforms, "platform")
			if err != nil {
				return nil, err
			}
			req.PlatformIDs = ids
		}
	}

	if d.Validate != nil {
		if err := d.Validate(req); err != nil {
			return nil, err
		}
	}

	hasLinks := req.LinkWeb != "" || req.LinkAndroid != "" || req.LinkIos != ""
	if req.ID == "" && hasLinks {
		// Creating a game does not store its links (CreateGameParams has no
		// link columns), so create a draft first, then update it with the
		// same fields so links are saved, and publish on that second call.
		created, err := d.Games.Save(ctx, req, false)
		if err != nil {
			return nil, err
		}
		req.ID = created.ID
	}

	saved, err := d.Games.Save(ctx, req, a.Publish)
	if err != nil {
		return nil, err
	}
	return saved, nil
}

func uploadFromURL(ctx context.Context, d Deps, raw json.RawMessage) (any, error) {
	var a struct {
		ImageURL string `json:"image_url"`
		Kind     string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &a); err != nil || a.ImageURL == "" {
		return nil, errors.New("image_url is required")
	}
	prefix := "banners"
	switch a.Kind {
	case "", "banner":
	case "screenshot":
		prefix = "screenshots"
	default:
		return nil, errors.New("kind must be banner or screenshot")
	}
	if d.Media == nil {
		return nil, errors.New("file uploads are not configured on this environment")
	}

	body, contentType, err := d.Fetcher.Fetch(ctx, a.ImageURL)
	if err != nil {
		return nil, err
	}
	url, err := d.Media.PutObject(ctx, prefix, bytes.NewReader(body), int64(len(body)), contentType)
	if err != nil {
		return nil, fmt.Errorf("store image: %w", err)
	}
	return map[string]any{"url": url, "content_type": contentType, "bytes": len(body)}, nil
}

// toRequest turns a stored game back into a save request.
func toRequest(g gamedto.GameResponse) gamedto.CreateOrUpdateGameRequest {
	req := gamedto.CreateOrUpdateGameRequest{
		ID:          g.ID,
		Slug:        g.Slug,
		Name:        g.Name,
		Dev:         g.Dev,
		Pitch:       g.Pitch,
		Description: g.Description,
		LinkWeb:     g.LinkWeb,
		LinkAndroid: g.LinkAndroid,
		LinkIos:     g.LinkIos,
		Bg:          g.Bg,
		Banner:      g.Banner,
		PlayStyle:   g.PlayStyle,
		Players:     g.Players,
		Tags:        g.Tags,
		VideoLinks:  g.VideoLinks,
		Screenshots: g.Screenshots,
	}
	for _, o := range g.Genres {
		req.GenreIDs = append(req.GenreIDs, o.ID)
	}
	for _, o := range g.Platforms {
		req.PlatformIDs = append(req.PlatformIDs, o.ID)
	}
	return req
}

// resolveOptions maps names, slugs or ids (case-insensitive) to option ids.
func resolveOptions(wanted []string, opts []gamedto.OptionItem, kind string) ([]string, error) {
	norm := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.TrimSuffix(s, " game")
		return strings.NewReplacer(" ", "", "-", "", "_", "", "&", "and").Replace(s)
	}
	ids := make([]string, 0, len(wanted))
	seen := map[string]bool{}
	for _, w := range wanted {
		key := norm(w)
		found := ""
		for _, o := range opts {
			if o.ID == w || norm(o.Name) == key || norm(o.Slug) == key {
				found = o.ID
				break
			}
		}
		if found == "" {
			names := make([]string, 0, len(opts))
			for _, o := range opts {
				names = append(names, o.Name)
			}
			return nil, fmt.Errorf("unknown %s %q; valid: %s", kind, w, strings.Join(names, ", "))
		}
		if !seen[found] {
			seen[found] = true
			ids = append(ids, found)
		}
	}
	return ids, nil
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify makes a URL slug from a game name, e.g. "Lumivara Online" ->
// "lumivara-online". Names with no Latin letters or digits get a time-based
// slug so creation never fails on an empty slug.
func Slugify(name string) string {
	s := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(s) > 80 {
		s = strings.TrimRight(s[:80], "-")
	}
	if len(s) < 2 {
		s = fmt.Sprintf("game-%d", time.Now().Unix())
	}
	return s
}

type gameSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Status  string `json:"status"`
	Pitch   string `json:"pitch"`
	Banner  string `json:"banner,omitempty"`
	LinkWeb string `json:"link_web,omitempty"`
}

func summarize(g gamedto.GameResponse) gameSummary {
	return gameSummary{ID: g.ID, Name: g.Name, Slug: g.Slug, Status: g.Status,
		Pitch: g.Pitch, Banner: g.Banner, LinkWeb: g.LinkWeb}
}

func prop(typ, desc string) map[string]any {
	return map[string]any{"type": typ, "description": desc}
}

func objectSchema(props map[string]any, required ...string) map[string]any {
	if props == nil {
		props = map[string]any{}
	}
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func stringArray(desc string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
}

func saveGameSchema() map[string]any {
	return objectSchema(map[string]any{
		"id":           prop("string", "Game id to update. Omit to create a new game."),
		"name":         prop("string", "Game name (2-255 chars). Required when creating."),
		"slug":         prop("string", "URL slug for a new game. Generated from the name when omitted."),
		"pitch":        prop("string", "One-line pitch shown on game cards (10-160 chars). Required when creating."),
		"description":  prop("string", "Description, up to 2000 chars. HTML like the site editor makes (<p>, <h3>, <ul><li>, <strong>, <u>) or plain text. Required when creating."),
		"dev":          prop("string", "Developer/creator name. Defaults to the account's display name."),
		"link_web":     prop("string", "https URL where the game can be played in a browser."),
		"link_android": prop("string", "Google Play URL."),
		"link_ios":     prop("string", "App Store URL."),
		"banner":       prop("string", "Banner image URL (16:9) from upload_image_from_url."),
		"screenshots":  stringArray("Screenshot URLs from upload_image_from_url (up to 6). Replaces the whole list."),
		"video_links":  stringArray("YouTube or Google Drive trailer links (up to 3). Replaces the whole list."),
		"genres":       stringArray("1-3 genre names or slugs, e.g. [\"MMORPG\", \"Adventure\"]. See list_options."),
		"platforms":    stringArray("Platform names or slugs, e.g. [\"Web game\", \"Mobile game\", \"PC game\"]."),
		"genre_ids":    stringArray("Genre ids instead of names."),
		"platform_ids": stringArray("Platform ids instead of names."),
		"play_style": map[string]any{
			"type":        "string",
			"enum":        []string{"casual", "co-op", "pvp", "pve", "story-driven", "fast-reflex", "brain"},
			"description": "One play style.",
		},
		"players": map[string]any{
			"type":        "string",
			"enum":        []string{"single-player", "local-2p", "online-multiplayer"},
			"description": "Number of players.",
		},
		"tags":    stringArray("Up to 5 extra tags, e.g. [\"pixel-art\"]. Replaces the whole list."),
		"publish": map[string]any{"type": "boolean", "description": "true to publish; default false saves a draft."},
	})
}

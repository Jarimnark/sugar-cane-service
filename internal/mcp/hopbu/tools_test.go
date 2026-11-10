package hopbu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Jarimnark/sugar-cane-service/internal/mcp"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

type saveCall struct {
	req     gamedto.CreateOrUpdateGameRequest
	publish bool
}

type fakeGames struct {
	stored map[string]gamedto.GameResponse
	saves  []saveCall
}

func (f *fakeGames) Options(context.Context) (*gamedto.MasterOptionResponse, error) {
	return &gamedto.MasterOptionResponse{
		Genres: []gamedto.OptionItem{
			{ID: "g-mmo", Name: "MMORPG", Slug: "mmorpg"},
			{ID: "g-adv", Name: "Adventure", Slug: "adventure"},
			{ID: "g-idle", Name: "Idle & Clicker", Slug: "idle-clicker"},
		},
		Platforms: []gamedto.OptionItem{
			{ID: "p-web", Name: "Web", Slug: "web"},
			{ID: "p-mob", Name: "Mobile", Slug: "mobile"},
		},
	}, nil
}

func (f *fakeGames) ListMine(context.Context) ([]gamedto.GameResponse, error) {
	out := []gamedto.GameResponse{}
	for _, g := range f.stored {
		out = append(out, g)
	}
	return out, nil
}

func (f *fakeGames) GetMine(_ context.Context, id string) (*gamedto.GameResponse, error) {
	g, ok := f.stored[id]
	if !ok {
		return nil, errors.New("game not found")
	}
	return &g, nil
}

func (f *fakeGames) Save(_ context.Context, req gamedto.CreateOrUpdateGameRequest, publish bool) (*gamedto.GameResponse, error) {
	f.saves = append(f.saves, saveCall{req, publish})
	id := req.ID
	if id == "" {
		id = "new-id"
	}
	return &gamedto.GameResponse{ID: id, Name: req.Name, Slug: req.Slug, Status: "draft"}, nil
}

type fakeFetcher struct {
	body []byte
	ct   string
	err  error
}

func (f fakeFetcher) Fetch(context.Context, string) ([]byte, string, error) {
	return f.body, f.ct, f.err
}

type fakeMedia struct {
	prefix string
	got    []byte
	ct     string
}

func (m *fakeMedia) PutObject(_ context.Context, prefix string, r io.Reader, size int64, ct string) (string, error) {
	b, _ := io.ReadAll(r)
	if int64(len(b)) != size {
		return "", errors.New("size mismatch")
	}
	m.prefix, m.got, m.ct = prefix, b, ct
	return "http://hopbu/media/" + prefix + "/x", nil
}

func tool(t *testing.T, tools []mcp.Tool, name string) mcp.Tool {
	t.Helper()
	for _, tl := range tools {
		if tl.Name == name {
			return tl
		}
	}
	t.Fatalf("tool %s not found", name)
	return mcp.Tool{}
}

func call(t *testing.T, d Deps, name, args string) (any, error) {
	t.Helper()
	return tool(t, Tools(d), name).Handler(context.Background(), json.RawMessage(args))
}

func TestSaveNewGameDraft(t *testing.T) {
	g := &fakeGames{}
	validated := false
	d := Deps{Games: g, Validate: func(any) error { validated = true; return nil }}

	_, err := call(t, d, "save_game", `{"name":"Lumivara Online!","pitch":"A pixel MMORPG","description":"<p>Hi there friends</p>","genres":["mmorpg","Adventure"],"platforms":["Web game"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !validated {
		t.Error("validate not called")
	}
	if len(g.saves) != 1 {
		t.Fatalf("saves = %d, want 1", len(g.saves))
	}
	s := g.saves[0]
	if s.publish || s.req.ID != "" || s.req.Slug != "lumivara-online" {
		t.Errorf("save = %+v", s)
	}
	if strings.Join(s.req.GenreIDs, ",") != "g-mmo,g-adv" || strings.Join(s.req.PlatformIDs, ",") != "p-web" {
		t.Errorf("ids = %v %v", s.req.GenreIDs, s.req.PlatformIDs)
	}
}

func TestSaveNewGameWithLinksSavesTwice(t *testing.T) {
	g := &fakeGames{}
	_, err := call(t, Deps{Games: g}, "save_game", `{"name":"Tide","pitch":"A fast game!!","description":"long enough","link_web":"https://tide.example/play","publish":true}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.saves) != 2 {
		t.Fatalf("saves = %d, want 2", len(g.saves))
	}
	if g.saves[0].publish || g.saves[0].req.ID != "" {
		t.Errorf("first save should be an unpublished create: %+v", g.saves[0])
	}
	if !g.saves[1].publish || g.saves[1].req.ID != "new-id" || g.saves[1].req.LinkWeb != "https://tide.example/play" {
		t.Errorf("second save should update links and publish: %+v", g.saves[1])
	}
}

func TestSaveUpdateMergesFields(t *testing.T) {
	g := &fakeGames{stored: map[string]gamedto.GameResponse{
		"g1": {
			ID: "g1", Slug: "tide", Name: "Tide", Pitch: "old pitch here", Description: "old description",
			LinkWeb: "https://tide.example", Banner: "http://hopbu/b.jpg", PlayStyle: "pve",
			Tags: []string{"a"}, Genres: []gamedto.OptionItem{{ID: "g-adv"}}, Platforms: []gamedto.OptionItem{{ID: "p-web"}},
		},
	}}
	_, err := call(t, Deps{Games: g}, "save_game", `{"id":"g1","pitch":"brand new pitch","tags":["x","y"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.saves) != 1 {
		t.Fatalf("saves = %d", len(g.saves))
	}
	r := g.saves[0].req
	if r.ID != "g1" || r.Name != "Tide" || r.Pitch != "brand new pitch" || r.Description != "old description" {
		t.Errorf("merged req = %+v", r)
	}
	if r.LinkWeb != "https://tide.example" || r.Banner != "http://hopbu/b.jpg" || r.PlayStyle != "pve" {
		t.Errorf("kept fields lost: %+v", r)
	}
	if strings.Join(r.Tags, ",") != "x,y" || strings.Join(r.GenreIDs, ",") != "g-adv" || strings.Join(r.PlatformIDs, ",") != "p-web" {
		t.Errorf("lists = %v %v %v", r.Tags, r.GenreIDs, r.PlatformIDs)
	}
}

func TestSaveErrors(t *testing.T) {
	g := &fakeGames{}
	if _, err := call(t, Deps{Games: g}, "save_game", `{"name":"X","genres":["Racing"]}`); err == nil || !strings.Contains(err.Error(), "unknown genre") {
		t.Errorf("want unknown genre error, got %v", err)
	}
	if _, err := call(t, Deps{Games: g}, "save_game", `{"id":"missing","pitch":"zzzzzzzzzz"}`); err == nil {
		t.Error("want error for unknown id")
	}
	v := Deps{Games: g, Validate: func(any) error { return errors.New("pitch too short") }}
	if _, err := call(t, v, "save_game", `{"name":"Tide"}`); err == nil || err.Error() != "pitch too short" {
		t.Errorf("validation error not returned: %v", err)
	}
	if len(g.saves) != 0 {
		t.Errorf("nothing should be saved on errors, got %d", len(g.saves))
	}
}

func TestUploadFromURL(t *testing.T) {
	m := &fakeMedia{}
	d := Deps{Media: m, Fetcher: fakeFetcher{body: []byte("img"), ct: "image/webp"}}

	out, err := call(t, d, "upload_image_from_url", `{"image_url":"https://x/y.webp","kind":"screenshot"}`)
	if err != nil {
		t.Fatal(err)
	}
	if m.prefix != "screenshots" || !bytes.Equal(m.got, []byte("img")) || m.ct != "image/webp" {
		t.Errorf("stored %q %q %q", m.prefix, m.got, m.ct)
	}
	if out.(map[string]any)["url"] != "http://hopbu/media/screenshots/x" {
		t.Errorf("out = %v", out)
	}

	if _, err := call(t, d, "upload_image_from_url", `{"image_url":"https://x/y.webp"}`); err != nil || m.prefix != "banners" {
		t.Errorf("default kind should be banner: %v %q", err, m.prefix)
	}
	if _, err := call(t, d, "upload_image_from_url", `{"image_url":"https://x","kind":"logo"}`); err == nil {
		t.Error("want error for bad kind")
	}
	if _, err := call(t, Deps{Fetcher: d.Fetcher}, "upload_image_from_url", `{"image_url":"https://x"}`); err == nil {
		t.Error("want error when storage is not configured")
	}
	bad := Deps{Media: m, Fetcher: fakeFetcher{err: errors.New("too big")}}
	if _, err := call(t, bad, "upload_image_from_url", `{"image_url":"https://x"}`); err == nil || err.Error() != "too big" {
		t.Errorf("fetch error not returned: %v", err)
	}
}

func TestListMyGamesIsCompact(t *testing.T) {
	g := &fakeGames{stored: map[string]gamedto.GameResponse{"g1": {ID: "g1", Name: "Tide", Description: "long"}}}
	out, err := call(t, Deps{Games: g}, "list_my_games", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(out)
	if strings.Contains(string(b), "description") || !strings.Contains(string(b), `"id":"g1"`) {
		t.Errorf("summary = %s", b)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Lumivara Online":        "lumivara-online",
		"  Tide--Bound! 2 ":      "tide-bound-2",
		"Idle & Clicker: Deluxe": "idle-clicker-deluxe",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
	if got := Slugify("เกมไทย"); !strings.HasPrefix(got, "game-") {
		t.Errorf("non-Latin name slug = %q", got)
	}
}

func TestResolveOptionsAcceptsIDsAndAmpersand(t *testing.T) {
	opts, _ := (&fakeGames{}).Options(context.Background())
	ids, err := resolveOptions([]string{"idle and clicker", "g-mmo", "MMORPG"}, opts.Genres, "genre")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "g-idle,g-mmo" {
		t.Errorf("ids = %v", ids)
	}
}

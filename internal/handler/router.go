package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	authhandler "github.com/Jarimnark/sugar-cane-service/internal/handler/auth"
	commenthandler "github.com/Jarimnark/sugar-cane-service/internal/handler/comment"
	devhandler "github.com/Jarimnark/sugar-cane-service/internal/handler/dev"
	gamehandler "github.com/Jarimnark/sugar-cane-service/internal/handler/game"
	listhandler "github.com/Jarimnark/sugar-cane-service/internal/handler/list"
	settinghandler "github.com/Jarimnark/sugar-cane-service/internal/handler/setting"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
)

func NewRouter(
	auth *authhandler.Handler,
	setting *settinghandler.Handler,
	games *gamehandler.Handler,
	list *listhandler.Handler,
	dev *devhandler.Handler,
	comments *commenthandler.Handler,
	sessions sessionrepo.Repository,
	mcp http.Handler, // nil = MCP endpoint disabled
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(appmw.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", handleHealthz)

	// MCP (Model Context Protocol) endpoint for AI clients. Authenticated by
	// its own bearer token, not by session cookies. Public URL: /api/mcp.
	if mcp != nil {
		r.Handle("/mcp", mcp)
	}

	r.Route("/api/v1", func(r chi.Router) {
		auth.MountPublic(r)
		games.MountPublic(r)

		r.Group(func(r chi.Router) {
			r.Use(appmw.OptionalAuth(sessions))
			comments.MountPublic(r)
			comments.MountAuthed(r)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmw.Auth(sessions))
			auth.MountAuthed(r)
			setting.MountAuthed(r)
			list.MountAuthed(r)
			dev.MountAuthed(r)
		})
	})

	return r
}

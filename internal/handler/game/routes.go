package gamehandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

func (h *Handler) MountPublic(r chi.Router) {
	r.Get("/home", h.Home)
	r.Get("/master-option", h.MasterOption)
	r.Get("/category-detail", h.CategoryDetail)
	r.Post("/game/list", h.List)
	r.Get("/game/suggest", h.Suggest)
	r.Get("/game/{id}", h.GetGame)
	r.Get("/game/{id}/related", h.RelatedGames)
	r.Post("/game/{id}/event", h.RecordEvent)
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Home(r.Context())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) MasterOption(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.MasterOption(r.Context())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) CategoryDetail(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.CategoryDetail(r.Context())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	req, err := decode[gamedto.ListRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	resp, err := h.svc.List(r.Context(), req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) Suggest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	resp, err := h.svc.Suggest(r.Context(), q)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.svc.GetGame(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) RelatedGames(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.svc.RelatedGames(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) RecordEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	eventType := r.URL.Query().Get("type")
	if err := h.svc.RecordEvent(r.Context(), id, eventType); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

package listhandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
)

func (h *Handler) MountAuthed(r chi.Router) {
	r.Put("/game/{id}/like", h.LikeGame)
	r.Delete("/game/{id}/like", h.UnlikeGame)
	r.Get("/me/list", h.GetMyList)
	r.Put("/me/list/{id}", h.AddToList)
	r.Delete("/me/list/{id}", h.RemoveFromList)
}

func (h *Handler) LikeGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.LikeGame(r.Context(), userID, id); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) UnlikeGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.UnlikeGame(r.Context(), userID, id); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) GetMyList(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	sort := r.URL.Query().Get("sort")
	resp, err := h.svc.GetMyList(r.Context(), userID, sort)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) AddToList(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.AddToList(r.Context(), userID, id); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) RemoveFromList(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.RemoveFromList(r.Context(), userID, id); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

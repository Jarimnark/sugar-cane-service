package devhandler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
	gamedto "github.com/Jarimnark/sugar-cane-service/pkg/dto/game"
)

func (h *Handler) MountAuthed(r chi.Router) {
	r.Get("/me/game", h.ListMyGames)
	r.Post("/me/game", h.CreateOrUpdateGame)
	r.Get("/me/game/{id}", h.GetMyGame)
	r.Delete("/me/game/{id}", h.DeleteMyGame)
}

func (h *Handler) ListMyGames(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	resp, err := h.svc.ListMyGames(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) CreateOrUpdateGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	publish, _ := strconv.ParseBool(r.URL.Query().Get("publish"))
	req, err := decode[gamedto.CreateOrUpdateGameRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	resp, err := h.svc.CreateOrUpdateGame(r.Context(), userID, req, publish)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) GetMyGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	resp, err := h.svc.GetMyGame(r.Context(), userID, id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) DeleteMyGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteMyGame(r.Context(), userID, id); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

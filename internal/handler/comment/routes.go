package commenthandler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
	commentdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/comment"
)

func (h *Handler) MountPublic(r chi.Router) {
	r.Get("/game/{id}/comments", h.ListComments)
}

func (h *Handler) MountAuthed(r chi.Router) {
	r.Post("/game/{id}/comments", h.CreateComment)
	r.Delete("/comment/{id}", h.DeleteComment)
	r.Put("/comment/{id}/like", h.LikeComment)
	r.Delete("/comment/{id}/like", h.UnlikeComment)
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	q := r.URL.Query()
	sort := q.Get("sort")
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	var viewerID pgtype.UUID
	if uid, ok := appmw.UserIDFromCtx(r.Context()); ok {
		viewerID = uid
	}

	resp, err := h.svc.List(r.Context(), gameID, sort, page, limit, viewerID)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var userID pgtype.UUID
	if uid, ok := appmw.UserIDFromCtx(r.Context()); ok {
		userID = uid
	}
	gameID := chi.URLParam(r, "id")
	req, err := decode[commentdto.CreateCommentRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	if !userID.Valid {
		req.Anonymous = true
	}
	resp, err := h.svc.Create(r.Context(), gameID, userID, req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusCreated, resp)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) LikeComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Like(r.Context(), id, userID); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) UnlikeComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Unlike(r.Context(), id, userID); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

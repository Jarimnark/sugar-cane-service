package settinghandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

func (h *Handler) MountAuthed(r chi.Router) {
	r.Post("/me/profile", h.CreateProfile)
	r.Patch("/me", h.UpdateMe)
	r.Post("/me/connection/{provider}", h.AddConnection)
	r.Delete("/me/connection/{provider}", h.RemoveConnection)
	r.Delete("/me", h.DeleteAccount)
}

func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	req, err := decode[userdto.UpdateProfileRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	resp, err := h.svc.CreateProfile(r.Context(), userID, req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusCreated, resp)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	req, err := decode[userdto.UpdateProfileRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	resp, err := h.svc.UpdateMe(r.Context(), userID, req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) AddConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	provider := chi.URLParam(r, "provider")
	if err := h.svc.AddConnection(r.Context(), userID, provider); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) RemoveConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	provider := chi.URLParam(r, "provider")
	if err := h.svc.RemoveConnection(r.Context(), userID, provider); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	if err := h.svc.DeleteAccount(r.Context(), userID); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusNoContent, nil)
}

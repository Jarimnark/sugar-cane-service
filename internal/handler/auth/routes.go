package authhandler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
)

func (h *Handler) MountPublic(r chi.Router) {
	r.Get("/auth/{provider}/start", h.Start)
	r.Get("/auth/{provider}/callback", h.Callback)
}

func (h *Handler) MountAuthed(r chi.Router) {
	r.Get("/auth/me", h.Me)
	r.Post("/auth/logout", h.Logout)
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		httputil.RespondError(w, domain.Internal(err))
		return
	}
	state := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	redirectURL, err := h.svc.OAuthStart(r.Context(), provider, state)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		httputil.RespondError(w, domain.BadRequest("invalid OAuth state", nil))
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", Path: "/", MaxAge: -1})
	out, err := h.svc.OAuthCallback(r.Context(), provider, r.URL.Query().Get("code"), r.RemoteAddr, r.UserAgent())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	h.setSessionCookie(w, out.SessionID, out.ExpiresAt)
	http.Redirect(w, r, h.cfg.FrontendURL, http.StatusFound)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	resp, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(appmw.SessionCookieName)
	if err != nil {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
		httputil.RespondError(w, err)
		return
	}
	h.clearSessionCookie(w)
	httputil.Respond(w, http.StatusNoContent, nil)
}

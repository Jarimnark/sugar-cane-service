package userhandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Jarimnark/sugar-cane-service/internal/config"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
	usersvc "github.com/Jarimnark/sugar-cane-service/internal/service/user"
)

var validate = validator.New()

type Handler struct {
	svc usersvc.Service
	cfg *config.Config
}

func New(svc usersvc.Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     appmw.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     appmw.SessionCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
	})
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, domain.BadRequest("invalid request body", err)
	}
	if err := validate.Struct(v); err != nil {
		return v, domain.BadRequest(err.Error(), err)
	}
	return v, nil
}

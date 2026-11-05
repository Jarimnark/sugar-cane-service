package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

const SessionCookieName = "session_id"

type sessionKey struct{}

func Auth(sessions sessionrepo.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				httputil.RespondError(w, domain.Unauthorized("authentication required"))
				return
			}

			session, err := sessions.Get(r.Context(), cookie.Value)
			if err != nil {
				httputil.RespondError(w, domain.Unauthorized("invalid or expired session"))
				return
			}

			ctx := context.WithValue(r.Context(), sessionKey{}, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth attaches the session to the request context when a valid
// cookie is present, but never rejects the request. Public endpoints that
// want to personalize the response (e.g. "liked" or "is_own" on comments)
// mount this instead of Auth.
func OptionalAuth(sessions sessionrepo.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			session, err := sessions.Get(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), sessionKey{}, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SessionFromCtx(ctx context.Context) (sqlcgen.Session, bool) {
	s, ok := ctx.Value(sessionKey{}).(sqlcgen.Session)
	return s, ok
}

func UserIDFromCtx(ctx context.Context) (pgtype.UUID, bool) {
	s, ok := ctx.Value(sessionKey{}).(sqlcgen.Session)
	if !ok {
		return pgtype.UUID{}, false
	}
	return s.UserID, true
}

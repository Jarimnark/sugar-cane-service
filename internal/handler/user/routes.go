package userhandler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

func (h *Handler) Mount(r chi.Router) {
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
	r.Get("/auth/oauth/{provider}", h.oauthRedirect)
	r.Get("/auth/oauth/{provider}/callback", h.oauthCallback)
}

func (h *Handler) MountAuthed(r chi.Router) {
	r.Post("/auth/logout", h.logout)
	r.Get("/users/me", h.me)
	r.Patch("/users/me", h.updateProfile)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	req, err := decode[userdto.RegisterRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusCreated, resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	req, err := decode[userdto.LoginRequest](r)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	out, err := h.svc.Login(r.Context(), req, r.RemoteAddr, r.UserAgent())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	h.setSessionCookie(w, out.SessionID, out.ExpiresAt)
	httputil.Respond(w, http.StatusOK, out.User)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
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
	resp, err := h.svc.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.Respond(w, http.StatusOK, resp)
}

// --- OAuth handlers ---

func (h *Handler) oauthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	cfg, err := h.oauthConfig(provider)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

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

	http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
}

func (h *Handler) oauthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		httputil.RespondError(w, domain.BadRequest("invalid OAuth state", nil))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
	})

	cfg, err := h.oauthConfig(provider)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	token, err := cfg.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		httputil.RespondError(w, domain.BadRequest("OAuth code exchange failed", err))
		return
	}

	providerID, email, displayName, photoURL, err := fetchProviderUser(r.Context(), provider, token, cfg)
	if err != nil {
		httputil.RespondError(w, domain.Internal(err))
		return
	}

	out, err := h.svc.LoginWithOAuth(r.Context(), provider, providerID, email, displayName, photoURL, r.RemoteAddr, r.UserAgent())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	h.setSessionCookie(w, out.SessionID, out.ExpiresAt)
	http.Redirect(w, r, h.cfg.FrontendURL, http.StatusFound)
}

// oauthConfig returns the provider-specific oauth2 config or a 400 error.
func (h *Handler) oauthConfig(provider string) (*oauth2.Config, error) {
	callbackURL := fmt.Sprintf("%s/api/v1/auth/oauth/%s/callback", h.cfg.OAuthRedirectBase, provider)
	switch provider {
	case "google":
		if h.cfg.GoogleClientID == "" {
			return nil, domain.BadRequest("Google OAuth is not configured", nil)
		}
		return &oauth2.Config{
			ClientID:     h.cfg.GoogleClientID,
			ClientSecret: h.cfg.GoogleClientSecret,
			RedirectURL:  callbackURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}, nil
	case "facebook":
		if h.cfg.FacebookAppID == "" {
			return nil, domain.BadRequest("Facebook OAuth is not configured", nil)
		}
		return &oauth2.Config{
			ClientID:     h.cfg.FacebookAppID,
			ClientSecret: h.cfg.FacebookAppSecret,
			RedirectURL:  callbackURL,
			Scopes:       []string{"email", "public_profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://www.facebook.com/dialog/oauth",
				TokenURL: "https://graph.facebook.com/oauth/access_token",
			},
		}, nil
	default:
		return nil, domain.BadRequest(fmt.Sprintf("unknown OAuth provider: %s", provider), nil)
	}
}

// fetchProviderUser retrieves (id, email, displayName, photoURL) from the provider API.
func fetchProviderUser(ctx context.Context, provider string, token *oauth2.Token, cfg *oauth2.Config) (id, email, displayName, photoURL string, err error) {
	client := cfg.Client(ctx, token)
	switch provider {
	case "google":
		return fetchGoogleUser(client)
	case "facebook":
		return fetchFacebookUser(client)
	default:
		return "", "", "", "", fmt.Errorf("fetchProviderUser: unknown provider %s", provider)
	}
}

func fetchGoogleUser(client *http.Client) (id, email, displayName, photoURL string, err error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", "", "", fmt.Errorf("fetchGoogleUser: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", "", "", "", fmt.Errorf("fetchGoogleUser: parse response: %w", err)
	}
	return info.ID, info.Email, info.Name, info.Picture, nil
}

func fetchFacebookUser(client *http.Client) (id, email, displayName, photoURL string, err error) {
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture")
	if err != nil {
		return "", "", "", "", fmt.Errorf("fetchFacebookUser: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", "", "", "", fmt.Errorf("fetchFacebookUser: parse response: %w", err)
	}
	return info.ID, info.Email, info.Name, info.Picture.Data.URL, nil
}

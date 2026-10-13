package authsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/Jarimnark/sugar-cane-service/internal/config"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	usersvc "github.com/Jarimnark/sugar-cane-service/internal/service/user"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

type svc struct {
	users usersvc.Service
	cfg   *config.Config
}

func New(users usersvc.Service, cfg *config.Config) Service {
	return &svc{users: users, cfg: cfg}
}

func (s *svc) OAuthStart(_ context.Context, provider, state string) (string, error) {
	cfg, err := s.oauthConfig(provider)
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

func (s *svc) OAuthCallback(ctx context.Context, provider, code, ip, userAgent string) (*userdto.LoginOutput, error) {
	oauthCfg, err := s.oauthConfig(provider)
	if err != nil {
		return nil, err
	}
	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, domain.BadRequest("OAuth code exchange failed", err)
	}
	providerID, email, displayName, photoURL, err := fetchProviderUser(ctx, provider, token, oauthCfg)
	if err != nil {
		return nil, domain.Internal(fmt.Errorf("fetch provider user: %w", err))
	}
	return s.users.LoginWithOAuth(ctx, provider, providerID, email, displayName, photoURL, ip, userAgent)
}

func (s *svc) Me(ctx context.Context, userID pgtype.UUID) (*userdto.UserResponse, error) {
	return s.users.Me(ctx, userID)
}

func (s *svc) Logout(ctx context.Context, sessionID string) error {
	return s.users.Logout(ctx, sessionID)
}

// oauthConfig returns the provider-specific oauth2.Config or a 400 error.
func (s *svc) oauthConfig(provider string) (*oauth2.Config, error) {
	callbackURL := fmt.Sprintf("%s/api/v1/auth/%s/callback", s.cfg.OAuthRedirectBase, provider)
	switch provider {
	case "google":
		if s.cfg.GoogleClientID == "" {
			return nil, domain.BadRequest("Google OAuth is not configured", nil)
		}
		return &oauth2.Config{
			ClientID:     s.cfg.GoogleClientID,
			ClientSecret: s.cfg.GoogleClientSecret,
			RedirectURL:  callbackURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}, nil
	case "facebook":
		if s.cfg.FacebookAppID == "" {
			return nil, domain.BadRequest("Facebook OAuth is not configured", nil)
		}
		return &oauth2.Config{
			ClientID:     s.cfg.FacebookAppID,
			ClientSecret: s.cfg.FacebookAppSecret,
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

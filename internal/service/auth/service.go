package authsvc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

type Service interface {
	OAuthStart(ctx context.Context, provider, state string) (redirectURL string, err error)
	OAuthCallback(ctx context.Context, provider, code, ip, userAgent string) (*userdto.LoginOutput, error)
	Me(ctx context.Context, userID pgtype.UUID) (*userdto.UserResponse, error)
	Logout(ctx context.Context, sessionID string) error
}

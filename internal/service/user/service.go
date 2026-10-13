package usersvc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

type Service interface {
	Register(ctx context.Context, req userdto.RegisterRequest) (*userdto.UserResponse, error)
	Login(ctx context.Context, req userdto.LoginRequest, ip, userAgent string) (*userdto.LoginOutput, error)
	LoginWithOAuth(ctx context.Context, provider, providerID, email, displayName, photoURL, ip, userAgent string) (*userdto.LoginOutput, error)
	Logout(ctx context.Context, sessionID string) error
	Me(ctx context.Context, userID pgtype.UUID) (*userdto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID pgtype.UUID, req userdto.UpdateProfileRequest) (*userdto.UserResponse, error)
}

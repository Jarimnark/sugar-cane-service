package settingsvc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

type Service interface {
	CreateProfile(ctx context.Context, userID pgtype.UUID, req userdto.UpdateProfileRequest) (*userdto.UserResponse, error)
	UpdateMe(ctx context.Context, userID pgtype.UUID, req userdto.UpdateProfileRequest) (*userdto.UserResponse, error)
	AddConnection(ctx context.Context, userID pgtype.UUID, provider string) error
	RemoveConnection(ctx context.Context, userID pgtype.UUID, provider string) error
	DeleteAccount(ctx context.Context, userID pgtype.UUID) error
}

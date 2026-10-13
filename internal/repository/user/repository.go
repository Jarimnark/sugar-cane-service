package userrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
)

type Repository interface {
	Create(ctx context.Context, params sqlcgen.CreateUserParams) (sqlcgen.User, error)
	GetByID(ctx context.Context, id pgtype.UUID) (sqlcgen.User, error)
	GetByEmail(ctx context.Context, email string) (sqlcgen.User, error)
	UpdateProfile(ctx context.Context, params sqlcgen.UpdateUserProfileParams) (sqlcgen.User, error)
	VerifyEmail(ctx context.Context, id pgtype.UUID) error
	SoftDelete(ctx context.Context, id pgtype.UUID) error
}

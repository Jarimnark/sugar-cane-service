package settingsvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jarimnark/sugar-cane-service/internal/dbutil"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	oauthrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/oauth"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
	userrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/user"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

type svc struct {
	users    userrepo.Repository
	oauths   oauthrepo.Repository
	sessions sessionrepo.Repository
	pool     *pgxpool.Pool
}

// New constructs the setting service with all required dependencies.
func New(users userrepo.Repository, oauths oauthrepo.Repository, sessions sessionrepo.Repository, pool *pgxpool.Pool) Service {
	return &svc{users: users, oauths: oauths, sessions: sessions, pool: pool}
}

func (s *svc) CreateProfile(ctx context.Context, userID pgtype.UUID, req userdto.UpdateProfileRequest) (*userdto.UserResponse, error) {
	// First login profile setup is identical to a regular profile update.
	return s.UpdateMe(ctx, userID, req)
}

func (s *svc) UpdateMe(ctx context.Context, userID pgtype.UUID, req userdto.UpdateProfileRequest) (*userdto.UserResponse, error) {
	u, err := s.users.UpdateProfile(ctx, sqlcgen.UpdateUserProfileParams{
		ID:          userID,
		DisplayName: req.DisplayName,
		Bio:         pgtype.Text{String: req.Bio, Valid: req.Bio != ""},
		PhotoUrl:    pgtype.Text{}, // photo_url is not changed via this endpoint
	})
	if err != nil {
		return nil, fmt.Errorf("UpdateMe: %w", err)
	}
	return userToDTO(u), nil
}

func (s *svc) AddConnection(_ context.Context, _ pgtype.UUID, _ string) error {
	// Adding an OAuth connection requires initiating a new OAuth flow.
	// Redirect the user to /auth/{provider}/start instead.
	return domain.BadRequest("to add an OAuth connection, redirect the user to /auth/{provider}/start", nil)
}

func (s *svc) RemoveConnection(ctx context.Context, userID pgtype.UUID, provider string) error {
	if err := s.oauths.DeleteByUserAndProvider(ctx, userID, provider); err != nil {
		return fmt.Errorf("RemoveConnection: %w", err)
	}
	return nil
}

func (s *svc) DeleteAccount(ctx context.Context, userID pgtype.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAccount: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := dbutil.WithTx(ctx, tx)

	if err := s.sessions.DeleteAllForUser(txCtx, userID); err != nil {
		return fmt.Errorf("DeleteAccount: delete sessions: %w", err)
	}

	if err := s.users.SoftDelete(txCtx, userID); err != nil {
		return fmt.Errorf("DeleteAccount: soft delete user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("DeleteAccount: commit: %w", err)
	}
	return nil
}

// userToDTO converts a sqlcgen.User to the shared UserResponse DTO.
func userToDTO(u sqlcgen.User) *userdto.UserResponse {
	var id [16]byte = u.ID.Bytes
	return &userdto.UserResponse{
		ID:            fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:]),
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		Bio:           u.Bio.String,
		PhotoURL:      u.PhotoUrl.String,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Time,
	}
}

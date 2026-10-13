package usersvc

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"

	"github.com/Jarimnark/sugar-cane-service/internal/dbutil"
	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	oauthrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/oauth"
	sessionrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/session"
	userrepo "github.com/Jarimnark/sugar-cane-service/internal/repository/user"
	"github.com/Jarimnark/sugar-cane-service/internal/store/sqlcgen"
	userdto "github.com/Jarimnark/sugar-cane-service/pkg/dto/user"
)

const sessionTTL = 30 * 24 * time.Hour

type svc struct {
	users    userrepo.Repository
	sessions sessionrepo.Repository
	oauths   oauthrepo.Repository
	pool     *pgxpool.Pool
}

func New(users userrepo.Repository, sessions sessionrepo.Repository, oauths oauthrepo.Repository, pool *pgxpool.Pool) Service {
	return &svc{users: users, sessions: sessions, oauths: oauths, pool: pool}
}

func (s *svc) Register(ctx context.Context, req userdto.RegisterRequest) (*userdto.UserResponse, error) {
	_, err := s.users.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, domain.BadRequest("email already registered", nil)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("Register: check email: %w", err)
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("Register: hash password: %w", err)
	}

	user, err := s.users.Create(ctx, sqlcgen.CreateUserParams{
		Email:        req.Email,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
		DisplayName:  req.DisplayName,
	})
	if err != nil {
		return nil, fmt.Errorf("Register: create user: %w", err)
	}

	return toUserResponse(user), nil
}

func (s *svc) Login(ctx context.Context, req userdto.LoginRequest, ip, userAgent string) (*userdto.LoginOutput, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.Unauthorized("invalid credentials")
		}
		return nil, fmt.Errorf("Login: get user: %w", err)
	}

	if !user.PasswordHash.Valid {
		return nil, domain.BadRequest("account uses social login — please sign in with OAuth", nil)
	}

	if err := verifyPassword(req.Password, user.PasswordHash.String); err != nil {
		return nil, domain.Unauthorized("invalid credentials")
	}

	sessionID, expiresAt, err := s.createSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("Login: create session: %w", err)
	}

	return &userdto.LoginOutput{SessionID: sessionID, ExpiresAt: expiresAt, User: toUserResponse(user)}, nil
}

func (s *svc) LoginWithOAuth(ctx context.Context, provider, providerID, email, displayName, photoURL, ip, userAgent string) (*userdto.LoginOutput, error) {
	// Fast path: existing OAuth account.
	account, err := s.oauths.GetByProvider(ctx, provider, providerID)
	if err == nil {
		user, err := s.users.GetByID(ctx, account.UserID)
		if err != nil {
			return nil, fmt.Errorf("LoginWithOAuth: get user: %w", err)
		}
		return s.buildLoginOutput(ctx, user, ip, userAgent)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("LoginWithOAuth: get oauth account: %w", err)
	}

	// No OAuth account yet. Find or create the user.
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("LoginWithOAuth: get user by email: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		// New user — create atomically with the oauth account row.
		user, err = s.createOAuthUser(ctx, email, displayName, photoURL, provider, providerID)
		if err != nil {
			return nil, err
		}
	} else {
		// Existing user — just link the new provider.
		if _, err := s.oauths.Create(ctx, user.ID, provider, providerID); err != nil {
			return nil, fmt.Errorf("LoginWithOAuth: link oauth account: %w", err)
		}
	}

	return s.buildLoginOutput(ctx, user, ip, userAgent)
}

func (s *svc) Logout(ctx context.Context, sessionID string) error {
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("Logout: %w", err)
	}
	return nil
}

func (s *svc) Me(ctx context.Context, userID pgtype.UUID) (*userdto.UserResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		return nil, fmt.Errorf("Me: %w", err)
	}
	return toUserResponse(user), nil
}

func (s *svc) UpdateProfile(ctx context.Context, userID pgtype.UUID, req userdto.UpdateProfileRequest) (*userdto.UserResponse, error) {
	user, err := s.users.UpdateProfile(ctx, sqlcgen.UpdateUserProfileParams{
		ID:          userID,
		DisplayName: req.DisplayName,
		Bio:         pgtype.Text{String: req.Bio, Valid: req.Bio != ""},
		PhotoUrl:    pgtype.Text{},
	})
	if err != nil {
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}
	return toUserResponse(user), nil
}

// --- helpers ---

func (s *svc) buildLoginOutput(ctx context.Context, user sqlcgen.User, ip, userAgent string) (*userdto.LoginOutput, error) {
	sessionID, expiresAt, err := s.createSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("buildLoginOutput: create session: %w", err)
	}
	return &userdto.LoginOutput{SessionID: sessionID, ExpiresAt: expiresAt, User: toUserResponse(user)}, nil
}

func (s *svc) createSession(ctx context.Context, userID pgtype.UUID, ip, userAgent string) (string, time.Time, error) {
	sessionID := newSessionID()
	expiresAt := time.Now().Add(sessionTTL)

	_, err := s.sessions.Create(ctx, sqlcgen.CreateSessionParams{
		ID:        sessionID,
		UserID:    userID,
		IpAddress: nil,
		UserAgent: pgtype.Text{String: userAgent, Valid: userAgent != ""},
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return sessionID, expiresAt, nil
}

// createOAuthUser atomically creates a new user and links an oauth account.
func (s *svc) createOAuthUser(ctx context.Context, email, displayName, photoURL, provider, providerID string) (sqlcgen.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sqlcgen.User{}, fmt.Errorf("createOAuthUser: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := dbutil.WithTx(ctx, tx)

	photo := pgtype.Text{}
	if photoURL != "" {
		photo = pgtype.Text{String: photoURL, Valid: true}
	}

	user, err := s.users.Create(txCtx, sqlcgen.CreateUserParams{
		Email:        email,
		PasswordHash: pgtype.Text{},
		DisplayName:  displayName,
	})
	if err != nil {
		return sqlcgen.User{}, fmt.Errorf("createOAuthUser: create user: %w", err)
	}

	// Update photo_url if provided (CreateUser doesn't include it).
	if photo.Valid {
		user, err = s.users.UpdateProfile(txCtx, sqlcgen.UpdateUserProfileParams{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Bio:         user.Bio,
			PhotoUrl:    photo,
		})
		if err != nil {
			return sqlcgen.User{}, fmt.Errorf("createOAuthUser: set photo: %w", err)
		}
	}

	if _, err := s.oauths.Create(txCtx, user.ID, provider, providerID); err != nil {
		return sqlcgen.User{}, fmt.Errorf("createOAuthUser: link oauth: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlcgen.User{}, fmt.Errorf("createOAuthUser: commit: %w", err)
	}

	return user, nil
}

func toUserResponse(u sqlcgen.User) *userdto.UserResponse {
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

func newSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// argon2id parameters (OWASP recommended minimums).
const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

func verifyPassword(password, encoded string) error {
	var version int
	var memory, iterations uint32
	var parallelism uint8
	var saltB64, hashB64 string

	_, err := fmt.Sscanf(encoded, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s",
		&version, &memory, &iterations, &parallelism, &saltB64)
	if err != nil {
		return fmt.Errorf("verifyPassword: parse hash: %w", err)
	}
	// hashB64 is the last $ segment.
	for i := len(encoded) - 1; i >= 0; i-- {
		if encoded[i] == '$' {
			hashB64 = encoded[i+1:]
			break
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return fmt.Errorf("verifyPassword: decode salt: %w", err)
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return fmt.Errorf("verifyPassword: decode hash: %w", err)
	}

	candidate := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(storedHash)))
	if subtle.ConstantTimeCompare(candidate, storedHash) != 1 {
		return errors.New("password mismatch")
	}
	return nil
}

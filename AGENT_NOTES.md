# Auth Agent Notes

## authsvc.New() signature

```go
authsvc.New(users usersvc.Service, cfg *config.Config) authsvc.Service
```

## Imports needed in main.go

```go
import (
    oauthrepo  "github.com/Jarimnark/sugar-cane-service/internal/repository/oauth"
    userrepo   "github.com/Jarimnark/sugar-cane-service/internal/repository/user"
    authsvc    "github.com/Jarimnark/sugar-cane-service/internal/service/auth"
    usersvc    "github.com/Jarimnark/sugar-cane-service/internal/service/user"
)
```

## Wiring example

```go
q := sqlcgen.New(pool)

sessions := sessionrepo.New(q)
users    := userrepo.New(q)
oauths   := oauthrepo.New(q)

userService := usersvc.New(users, sessions, oauths, pool)
authService := authsvc.New(userService, cfg)
```

## Notes

- `main.go` was updated as part of this implementation because the `authsvc.New()` signature changed from `New()` to `New(users, cfg)` — a non-compiling skeleton is not acceptable.
- State cookie generation and validation live in the handler (`internal/handler/auth/routes.go`); `OAuthStart` and `OAuthCallback` in the service layer are stateless with respect to the CSRF token.
- `oauthConfig()`, `fetchProviderUser()`, `fetchGoogleUser()`, `fetchFacebookUser()` were moved verbatim from `internal/handler/user/routes.go` into `internal/service/auth/impl.go`. The handler copies in `internal/handler/user/routes.go` are intentionally left in place (they still serve the legacy `/auth/oauth/{provider}` routes).

---

# Setting Agent Notes

## settingsvc.New() signature
```go
settingsvc.New(users userrepo.Repository, oauths oauthrepo.Repository, sessions sessionrepo.Repository, pool *pgxpool.Pool) settingsvc.Service
```

## Imports needed in main.go
```go
settingsvc "github.com/Jarimnark/sugar-cane-service/internal/service/setting"
```
(already present; only the call site needed updating to pass deps)

## New migration
`00012_add_users_deleted_at.sql` — adds `deleted_at TIMESTAMPTZ NULL` to `users` table.
Run `goose up` before starting the server or running integration tests:
```bash
goose -dir db/migrations postgres "$DATABASE_URL" up
```

## What was implemented
- `CreateProfile` — delegates to `UpdateMe` (first OAuth login profile setup)
- `UpdateMe` — updates `display_name` and `bio` via `userrepo.UpdateProfile`; leaves `photo_url` unchanged
- `AddConnection` — returns HTTP 400 with hint to use `/auth/{provider}/start`
- `RemoveConnection` — deletes the OAuth account row for the given user+provider pair
- `DeleteAccount` — runs in a transaction: deletes all sessions, then soft-deletes the user

## New sqlcgen queries (from setting.sql)
- `DeleteOAuthAccountByUserAndProvider` — params struct: `DeleteOAuthAccountByUserAndProviderParams{UserID, Provider}`
- `SoftDeleteUser` — takes `pgtype.UUID` directly

## Soft-delete visibility
`GetUserByID` and `GetUserByEmail` in `users.sql` now filter `AND deleted_at IS NULL`,
so soft-deleted users are invisible to all existing callers automatically.

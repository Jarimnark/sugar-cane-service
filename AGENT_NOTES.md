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

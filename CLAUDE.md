# CLAUDE.md — sugar-cane-service

Go backend API for SugarCane, a web-game shelf platform. Pairs with `sugar-cane-web` (Next.js). Primary goal is a working product; secondary goal is hands-on infra practice (self-hosted VPS, Docker Compose, CI/CD, observability).

---

## Tech stack

| Area | Choice |
|---|---|
| Language | Go |
| Router | chi (`go-chi/chi`) |
| Database | PostgreSQL |
| DB access | pgx + sqlc (type-safe, parameterised queries) |
| Migrations | goose |
| Auth | Server-side sessions in Postgres, `httpOnly` cookie — **no JWT** |
| Password hashing | argon2id |
| File uploads | S3-compatible: MinIO (local), Cloudflare R2 (prod) |
| Logging | `log/slog` JSON to stdout |

---

## Project layout

```
cmd/
  api/
    main.go          ← entry point: wire deps, start server
internal/
  handler/           ← HTTP handlers (chi routes)
  service/           ← business logic
  store/             ← DB access (sqlc-generated + repository wrappers)
  domain/            ← shared types, AppError, constants
  middleware/        ← auth session, logging, rate-limit
  config/            ← env-var loading
db/
  migrations/        ← goose migration files
  queries/           ← sqlc .sql query files
  sqlc.yaml
docker-compose.yml   ← postgres + minio for local dev
```

---

## Error handling

Use a custom `AppError` type with an HTTP status code baked in. Never return bare `errors.New` from handlers.

```go
type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }
```

Handlers unwrap `*AppError` and write the status code + JSON body. Wrap internal errors with `fmt.Errorf("context: %w", err)` before returning.

---

## Dependency injection & testing

Wire all dependencies explicitly — no package-level globals. Every service and store takes its deps as constructor arguments (interfaces, not concrete types). This makes swapping test doubles straightforward without mocking the real DB unnecessarily.

For tests that touch the DB, use a real Postgres instance (via `docker compose`). Integration tests are preferred over mocks for the store layer. Unit tests with interfaces are fine for the service layer.

---

## Hard rules — never do these

1. **No JWT.** Auth is server-side sessions stored in Postgres. Session ID in an `httpOnly`, `Secure`, `SameSite=Lax` cookie only.
2. **No hardcoded values.** No `localhost`, ports, passwords, or secrets in code. Everything comes from env vars via `config/`.
3. **No raw SQL outside sqlc.** All queries go through sqlc-generated code. Write SQL in `db/queries/`, run `sqlc generate`, use the generated functions.
4. **No skipping goose migrations.** Schema changes must be a goose migration file. Never `ALTER TABLE` manually or in application startup code.
5. **No user IDs from the client.** Every write checks ownership server-side (`owner_id = session.user_id`). Ignore any user ID sent in the request body.

---

## Config — env vars only

All config is loaded at startup from environment variables. Fail fast with a clear error if a required var is missing. Never fall back to hardcoded defaults for secrets or URLs.

| Variable | Local example |
|---|---|
| `DATABASE_URL` | `postgres://app:app@localhost:5432/sugarcane?sslmode=disable` |
| `COOKIE_DOMAIN` | `localhost` |
| `COOKIE_SECURE` | `false` |
| `CORS_ORIGIN` | `http://localhost:3000` |
| `S3_ENDPOINT` | `http://localhost:9000` |
| `S3_BUCKET` | `sugarcane` |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | MinIO root creds |

---

## Local dev setup

Dependencies run in Docker; the Go app runs natively.

```bash
docker compose up -d          # start postgres + minio
goose -dir db/migrations postgres "$DATABASE_URL" up
go run ./cmd/api
```

`docker-compose.yml` defines `postgres` and `minio` services on a private network. The app service is **not** in compose — run it with `go run` or a local binary.

---

## Verification before done

Before declaring a task complete:

```bash
go build ./...
go vet ./...
go test ./...
```

All three must pass. If tests don't exist yet for a new piece of code, note that explicitly rather than claiming full coverage.

---

## Logging

Use `log/slog` with JSON output to stdout. Attach request-scoped fields (request ID, user ID when authed) via `slog.With`. Never use `fmt.Println` or `log.Printf` in application code.

---

## Chi conventions (note for Gin users)

The user knows Gin well — chi is the decided router. Key differences to keep in mind:

- Route params: `chi.URLParam(r, "id")` not `c.Param("id")`
- Middleware: standard `func(http.Handler) http.Handler` — use `r.Use(...)`
- No context helpers — use `r.Context()` and pass it down explicitly
- Response: write to `w http.ResponseWriter` directly or via a small `respond(w, status, v)` helper

---

## Migrations

Use goose with SQL migration files (not Go). File naming: `00001_create_users.sql`, `00002_create_sessions.sql`, etc.

Zero-downtime pattern for adding columns:
1. Add nullable column (one migration)
2. Backfill in a separate step
3. Add constraint (another migration)

---

## Security checklist (consult before shipping any endpoint)

- [ ] Password hashed with argon2id
- [ ] Rate limiting on login, signup, and password reset
- [ ] Email verified before user can publish content
- [ ] `owner_id` check on every write endpoint
- [ ] File uploads: validate MIME type + size; serve from separate domain/bucket
- [ ] Rich text sanitised with `bluemonday`
- [ ] CORS limited to `CORS_ORIGIN` env var; `SameSite=Lax` + origin check on writes
- [ ] Secrets only in env / CI secrets, never in git

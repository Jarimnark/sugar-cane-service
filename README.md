# sugar-cane-service

Backend API for **SugarCane**: a web-game shelf where registered users manage their own content, and the public site is rendered for SEO by [`sugar-cane-web`](https://github.com/Jarimnark/sugar-cane-web).

Secondary goal of this project: **hands-on infra practice** (self-hosted server, database, CI/CD, monitoring).

## Tech stack

| Area | Choice | Status |
| --- | --- | --- |
| Language | Go | Decided |
| HTTP router | chi (or Echo) | Decided: chi unless it gets in the way |
| Database | PostgreSQL | Decided |
| DB access | pgx + sqlc (type-safe, parameterized queries) | Decided |
| Migrations | goose | Decided |
| Auth | Server-side sessions in Postgres, `httpOnly` cookie | Decided |
| Password hashing | argon2id | Decided |
| File uploads | S3-compatible API: MinIO (local), Cloudflare R2 (prod) | Decided |
| Logging | `log/slog` JSON to stdout | Decided |
| Hosting | Hetzner VPS (Singapore) + Docker Compose + Caddy | Later (infra phase) |
| CI/CD | GitHub Actions → container registry → SSH deploy | Later (infra phase) |

## Architecture

```
Cloudflare (DNS, CDN, DDoS, TLS proxy)
        │
   Hetzner VPS
   └─ Docker Compose
      ├─ caddy     → reverse proxy, auto HTTPS
      │   ├─ example.com      → sugar-cane-web (Next.js SSR)
      │   └─ api.example.com  → sugar-cane-service (this repo)
      ├─ sugar-cane-service
      └─ postgres  (private network only, never exposed)
Cloudflare R2 → user uploads + DB backups
```

## Auth and authorization

- **Authentication:** email + password (argon2id), email verification, rate-limited login and signup.
- **Session:** random session ID in an `httpOnly`, `Secure`, `SameSite=Lax` cookie scoped to the parent domain; session rows live in Postgres, so logout and "log out all devices" are instant. No JWT: one web client and one backend don't need it.
- **Authorization:** every write checks ownership server-side (`owner_id = session.user_id`). Never trust a user ID sent by the client.

## Database practices

- Schema changes only through goose migrations, run in CI before the new container starts.
- Least-privilege roles: `app` (CRUD only), `migrator` (schema changes); the app never connects as `postgres`.
- User content stored as `jsonb` where flexible (GIN index), `owner_id` foreign keys everywhere.
- Zero-downtime habits: nullable column → backfill → constraint; `CREATE INDEX CONCURRENTLY`.

## Development principles (do from day one)

These keep the app deployable later without rework:

- **Dependencies in Docker locally:** Postgres and MinIO via `docker compose`; the Go app runs natively.
- **Config from env vars only:** DB URL, cookie domain, CORS origin, bucket, `Secure` flag. No hardcoded `localhost`.
- **Migrations from the first table.**
- **Uploads through the S3 interface** so MinIO → R2 is an env-var change.
- **Log to stdout** and expose **`/healthz`**.
- **Deploy a rough version early** (login + one editable page) to surface cookie, CORS and env issues while the app is small.

### Environment variables (planned)

| Variable | Example (local) |
| --- | --- |
| `DATABASE_URL` | `postgres://app:app@localhost:5432/sugarcane?sslmode=disable` |
| `COOKIE_DOMAIN` | `localhost` (prod: `.example.com`) |
| `COOKIE_SECURE` | `false` (prod: `true`) |
| `CORS_ORIGIN` | `http://localhost:3000` |
| `S3_ENDPOINT` / `S3_BUCKET` | `http://localhost:9000` / `sugarcane` |

## Running locally

The Go app runs natively on your host; Postgres and MinIO run in Docker.

### Prereqs

- Docker + Docker Compose
- Go toolchain
- `goose` and `sqlc`: `make install-tools`

### First-time setup

```bash
cp .env.example .env
# edit DATABASE_URL to use host port 6391 (compose maps 6391 → 5432):
#   postgres://app:app@localhost:6391/sugarcane?sslmode=disable

docker compose up -d           # start postgres + minio
make migrate-up                # run goose migrations
make run                       # start the API on :8080
```

The `make` targets auto-load `.env` (`set -a; . ./.env; set +a`) so you don't need to export vars yourself. Override with `make run ENV_FILE=.env.staging` to use a different file.

### MinIO bucket

The compose file sets `MINIO_DEFAULT_BUCKETS: sugarcane`, so the `sugarcane` bucket is auto-created on first start. The console is at `http://localhost:9001` (login: `minioadmin` / `minioadmin`).

### MinIO image note (important — don't swap it back)

We use `bitnamilegacy/minio:latest`, **not** `minio/minio` or `quay.io/minio/minio`. Context:

- `minio/minio` on Docker Hub was removed by MinIO Inc.
- `quay.io/minio/minio` (both `:latest` and pinned release tags) now requires a Quay login — MinIO is pushing users toward their paid Enterprise product.
- `bitnami/minio` was moved behind the paid "Bitnami Secure" tier after Broadcom's acquisition in 2025.
- `bitnamilegacy/*` is Bitnami's publicly-accessible archive namespace. Free, no auth, drop-in S3-compatible.

If `bitnamilegacy` is ever retired, the fallback is to create a free Quay account, `docker login quay.io`, and switch to `quay.io/minio/minio:RELEASE.<date>`.

Prod uses Cloudflare R2 anyway (same S3 API), so this only matters for local dev.

### Everyday commands

```bash
make run                       # start the API (loads .env)
make test                      # go test ./...
make lint                      # go build + go vet
make sqlc                      # regenerate sqlc after editing db/queries/*.sql
make migrate-up                # run new migrations
make migrate-status            # check migration state

docker compose ps              # see what's running
docker compose logs minio -f   # tail minio logs
docker compose stop            # stop (keeps data in volumes)
docker compose down -v         # nuke everything incl. volumes (destructive!)
```

### Connecting the frontend

The Next.js app (`sugar-cane-web`) runs on `http://localhost:3000`. The backend `.env` has `CORS_ORIGIN=http://localhost:3000` and `COOKIE_DOMAIN=localhost` / `COOKIE_SECURE=false` so session cookies work over plain HTTP in dev. The frontend must send requests with credentials included (`credentials: 'include'` for fetch, `withCredentials: true` for axios) or the session cookie is dropped.

### Pre-commit check

```bash
make lint && make test
```

## MCP endpoint (AI clients)

`POST /mcp` (public URL `/api/mcp` behind Caddy) is a Model Context Protocol
server so Claude and other MCP clients can manage games. It is stateless,
JSON-only Streamable HTTP, and **off unless `MCP_TOKEN` is set**.

| Variable | Meaning |
|---|---|
| `MCP_TOKEN` | Shared secret, at least 32 chars (`openssl rand -hex 32`). Clients send `Authorization: Bearer <token>`. |
| `MCP_USER_ID` | UUID of the Hopbu user the tools act as. Every write is owned by this user. |

Tools: `list_options`, `list_my_games`, `get_my_game`, `save_game`
(create, or update only the fields sent; draft unless `publish: true`;
genres/platforms by name), `upload_image_from_url` (server downloads a
JPEG/PNG/WebP up to 2 MB and stores it; refuses private/internal addresses).

Connect from Claude Code:

```bash
claude mcp add --transport http hopbu http://<host>/api/mcp \
  --header "Authorization: Bearer $MCP_TOKEN"
```

Code: `internal/mcp` (protocol, stdlib only), `internal/mcp/hopbu` (tools),
`cmd/api/mcp.go` (wiring).

## Security checklist

- [ ] Passwords hashed with argon2id
- [ ] Rate limiting on login, signup and password reset
- [ ] Email verification before publishing content
- [ ] Ownership check on every write endpoint
- [ ] Uploads: validate type and size, never execute, serve from a separate domain/bucket
- [ ] Sanitize stored rich text (bluemonday) as a second line behind the frontend
- [ ] CORS limited to the web origin; CSRF covered by `SameSite=Lax` + origin check on writes
- [ ] Secrets only in env / CI secrets, never in git

## Infra roadmap (after the app works)

1. **Go live:** harden the VPS (non-root user, SSH keys only, `ufw` 22/80/443, `fail2ban`, `unattended-upgrades`), Docker Compose, Caddy, `.env` readable by owner only.
2. **CI/CD:** test → build image → push to registry → SSH deploy (`docker compose pull && up -d`), migrations as a pipeline step.
3. **Data safety:** nightly `pg_dump` to R2 with retention; monthly restore drill. Later: point-in-time recovery with pgBackRest or WAL-G.
4. **Observability:** Uptime Kuma alerts; Prometheus + Grafana + node/postgres exporters; Loki for logs; alert on disk > 80%.
5. **Infra as code:** Terraform/OpenTofu for server, firewall and DNS; Ansible for hardening. Goal: destroy and rebuild the server from scratch, DB restored from backup, in about 30 minutes.

Later DB topics: PgBouncer, `pg_stat_statements` tuning, streaming replica + manual failover, major-version upgrade with `pg_upgrade`.

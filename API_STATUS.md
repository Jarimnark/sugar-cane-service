# API Status

> **Rule:** Update this file in the same session as any implementation change. Never leave status out of sync with the code.

Status values: `todo` · `inprogress` · `done` · `fix`

---

## Auth

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| GET | /auth/:provider/start | done | OAuth redirect; state cookie (MaxAge 300) |
| GET | /auth/:provider/callback | done | Validates state cookie, sets session cookie, redirects to FrontendURL |
| GET | /auth/me | done | Returns current user profile |
| POST | /auth/logout | done | Clears session cookie |

---

## Setting

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| POST | /me/profile | done | Delegates to UpdateMe (idempotent upsert) |
| PATCH | /me | done | Updates display name / bio |
| POST | /me/connection/:provider | done | Stub — returns BadRequest (use /auth/:provider/start) |
| DELETE | /me/connection/:provider | done | Unlinks OAuth provider |
| DELETE | /me | done | Soft-delete: deleted_at stamp + session purge (migration 00012) |

---

## Game (public)

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| GET | /home | done | trending (5 by visit_count), popular (6 by like_count), new (5 by created_at) |
| GET | /master-option | done | Genres + Platforms dropdown options |
| GET | /category-detail | done | Game counts by genre and platform |
| POST | /game/list | done | Filtered by genreID/platformID, sorted by popular/new, paginated |
| GET | /game/suggest | done | ILIKE search on name, returns top 10 by like_count |
| GET | /game/:id | done | Game detail; increments visit_count |
| GET | /game/:id/related | done | Related by tags overlap, excludes self |
| POST | /game/:id/event | done | Records visit event (increments visit_count) |

---

## My List

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| PUT | /game/:id/like | done | Atomic: INSERT game_like + UPDATE like_count in tx |
| DELETE | /game/:id/like | done | Atomic: DELETE game_like + GREATEST(like_count-1,0) in tx |
| GET | /me/list | done | Saved games ordered by saved_at DESC |
| PUT | /me/list/:id | done | Add to saved_games (ON CONFLICT DO NOTHING) |
| DELETE | /me/list/:id | done | Remove from saved_games |

---

## Dev

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| GET | /me/game | done | Lists games by owner |
| POST | /me/game | done | id="" → create; id=UUID → update; ?publish=true publishes |
| GET | /me/game/:id | done | Owner-only view; 403 if not owner |
| DELETE | /me/game/:id | done | Owner-enforced by WHERE owner_id=$2 in SQL |

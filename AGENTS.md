# AGENTS.md

## Commands

```bash
# Backend — run from repo root
cd backend && go build ./...            # compile check
cd backend && go vet ./...              # lint
cd backend && go test ./...             # all tests
cd backend && go test ./internal/application/service/...   # single package

# Frontend — run from repo root
cd frontend && npm run build            # typecheck + build (vue-tsc -b && vite build)
cd frontend && npx vue-tsc --noEmit     # typecheck only
cd frontend && npm run test:run         # tests (single run)
cd frontend && npm run test             # tests (watch)
cd frontend && npm run lint             # eslint with --fix
```

Verify order after edits: `go vet` → `go build` → `go test` (backend), `vue-tsc --noEmit` → `npm run build` (frontend).

## Architecture

Monorepo: `backend/` (Go) + `frontend/` (Vue 3). No shared code between them.

**Backend — 4-layer DDD (strict dependency direction)**:
- `internal/domain/entity/` — pure structs, zero external imports
- `internal/domain/repository/` — interfaces only
- `internal/application/service/` — business logic, depends on domain interfaces
- `internal/infrastructure/persistence/model/` — GORM structs with `ToEntity()` / `FromEntity()` bridging
- `internal/infrastructure/persistence/repository/` — GORM implementations of domain interfaces
- `internal/infrastructure/middleware/` — Gin middleware (auth, cors, pagination, space_auth)
- `internal/interfaces/handler/` — Gin handlers
- `pkg/` — shared utilities (jwt, hash, response, tree, diff, validator)

**Frontend**:
- `src/api/` — axios client + per-module API functions
- `src/stores/` — Pinia stores (auth, space, page)
- `src/views/` — page views by feature
- `src/components/editor/` — TipTap editor
- `@/` path alias → `src/`
- Element Plus auto-imported via `unplugin-auto-import` + `unplugin-vue-components`

## Key Conventions

- **API envelope**: all responses use `{ code, message, data }`. Success code is `0`. Paginated variant wraps data in `{ items, total, page, page_size }`.
- **Config**: Viper loads `backend/configs/config.yaml`. Env overrides use `MOSS_` prefix with `_` separators (e.g. `MOSS_DATABASE_HOST`). See `.env.example`.
- **Port**: backend runs on port **8033** in both local dev and Docker Compose. Frontend Vite proxy and Nginx both target `:8033`.
- **Tests**: backend uses `testify/assert` with hand-rolled mock repos in test files (no mockgen). Tests live alongside source (same package).
- **Migrations**: plain SQL files in `backend/migrations/001–005`. Loaded via `docker-entrypoint-initdb.d` in Docker; applied manually for local dev.
- **First user**: the first registered user is auto-promoted to `role=admin`.
- **No comments in code** per project convention — do not add comments unless asked.

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**MossBase** — 企业级 Wiki 知识库系统。苔藓（Moss）+ 知识库（Base），小而有机。

- Remote: `git@github.com:gachal/MossBase.git`
- License: MIT (Copyright 2026, gachal)

## Tech Stack

- **Backend**: Go (Gin + GORM) + MySQL 8.0 + JWT + Viper + Zap
- **Frontend**: Vue 3 + TypeScript + Pinia + Vue Router + Element Plus + TipTap v3 + Vite
- **Vector DB**: Qdrant (Phase 3)
- **Async Tasks**: Asynq + Redis (Phase 3)
- **MCP**: Official Go MCP SDK (Phase 4)
- **Deployment**: Docker Compose

## Architecture

**Backend**: 4-layer DDD (`interfaces -> application -> domain <- infrastructure`). Domain layer has zero external dependencies.

**Frontend**: Feature-organized with Pinia stores, auto-imported Element Plus components.

**Page Tree**: Adjacency list model (`parent_id` self-referential FK), loaded in one query and built in memory.

## Common Commands

```bash
# Backend
cd backend
go build ./...          # Build all
go test ./...           # Run all tests
go vet ./...            # Lint
go mod tidy             # Install deps

# Frontend
cd frontend
npm run dev             # Dev server (port 5173, proxies /api to :8033)
npm run build           # Production build
npm run test            # Run tests (Vitest)
npx vue-tsc --noEmit    # Type check

# Docker
docker-compose up       # Start all services (MySQL + Redis + backend + frontend)
```

## Project Structure

```
backend/
  cmd/server/main.go          # Entry point
  configs/config.yaml         # Viper config
  internal/
    domain/                   # Pure domain (entities + repo interfaces)
    application/              # DTOs + services
    infrastructure/           # GORM models, repo impls, middleware, config, DB
    interfaces/               # Gin handlers + router
  pkg/                        # Shared: jwt, hash, response, tree, diff
  migrations/                 # SQL migrations (001-005)

frontend/
  src/
    api/                      # Axios client + module APIs
    components/editor/        # TipTap editor
    components/page-tree/     # Page tree sidebar
    layouts/                  # MainLayout, AuthLayout, AdminLayout
    router/                   # Vue Router config
    stores/                   # Pinia stores (auth, space, page, version, admin)
    types/                    # TypeScript types
    utils/                    # tree.ts, storage.ts
    views/                    # Page views by feature
```

## Key Patterns

- All API responses use envelope: `{ code, message, data }` with paginated variant
- GORM models have `ToEntity()` / `FromEntity()` for domain-persistence boundary
- JWT auth via `middleware.Auth()`, admin check via `middleware.AdminAuth()`
- First registered user becomes system admin (role=admin)
- Space permissions: admin/member/viewer at space level

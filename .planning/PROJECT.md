# Podgrab

## What This Is

Podgrab is a self-hosted podcast manager that automatically downloads podcast episodes as they become live. Built in Go with a server-side rendered web UI (Gin + Vue 2 + SQLite), it lets users subscribe to podcasts, search iTunes/PodcastIndex/gPodder, download episodes, organize with tags, and play episodes in-browser with real-time sync across tabs.

The project is being modernized — upgrading outdated dependencies (Go 1.15, GORM v1, Vue 2) and replacing the server-rendered UI with a React SPA, done incrementally so all existing features keep working throughout.

## Core Value

Podcast episodes are automatically downloaded and available — that download-and-organize loop must never break.

## Requirements

### Validated

<!-- Extracted from existing codebase capabilities -->

- ✓ Subscribe to podcasts via RSS URL — existing
- ✓ Search and discover podcasts via iTunes, PodcastIndex, and gPodder — existing
- ✓ Automatically download new episodes on a schedule — existing
- ✓ Play episodes in-browser with WebSocket-synced player state — existing
- ✓ Organize podcasts with tags — existing
- ✓ Import/export subscriptions via OPML — existing
- ✓ Create database backups — existing
- ✓ Basic HTTP auth protection — existing
- ✓ Download episode files to configurable storage with NFO metadata — existing
- ✓ gPodder directory integration (search by tag, top lists) — existing

### Active

- [ ] Upgrade Go and all dependencies to current versions (Go 1.22+, GORM v1.25+, Gin latest, etc.)
- [ ] Replace Vue 2 frontend with a React SPA, built incrementally alongside existing UI
- [ ] Add proper REST API layer on the Go backend for the React frontend to consume
- [ ] Fix known bugs (typo handler, `removeStartingSlash`, concurrency batching, date parsing)
- [ ] Move hardcoded PodcastIndex API credentials to environment variables
- [ ] Modernize error handling (replace `fmt.Println` with structured logging, propagate errors properly)
- [ ] Add thread-safe state management for WebSocket maps (replace bare maps with `sync.Map` or mutex-protected)

### Out of Scope

- Mobile native app — web-first, responsive is sufficient for now
- Multi-user authentication / role-based access — single-user tool, basic auth is adequate
- Podcast streaming without download — Podgrab is a download-first manager
- Social features (comments, sharing) — not core to the podcast manager value proposition

## Context

- Brownfield project: existing Go codebase with 6,000+ lines across controllers, services, models, and DB layer
- Monolithic MVC architecture with server-side rendered templates (Go `html/template`) mixed with inline Vue 2
- No test coverage at all — no test files exist in the repo
- Global mutable state (`db.DB` singleton, WebSocket maps without mutex) makes testing and refactoring harder
- Controllers bypass service layer in some places, hitting `db.*` directly
- Dual logging system: `zap.SugaredLogger` (rarely used) and `fmt.Println` (pervasively used)
- Docker-based deployment with Alpine runtime, persistent volumes for `/config` and `/assets`
- Frontend assets are vendored JS/CSS libraries — no build system, no npm

## Constraints

- **Tech Stack**: Go backend (must stay Go), SQLite via GORM (must stay SQLite), React for new frontend
- **Incremental**: All existing features must work throughout the migration — no flag-day cutover
- **Compatibility**: Existing data (SQLite DB, downloaded files, OPML imports) must continue to work
- **Deployment**: Must remain Docker-deployable with the same volume mount pattern (`/config`, `/assets`)
- **No feature loss**: Every current capability (search, download, tags, OPML, gPodder, backups, player sync) must be preserved

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| React for new frontend | User chose React over Svelte/Angular for new SPA | — Pending |
| Incremental migration | Keep existing features working throughout, avoid big-bang cutover | — Pending |
| API + frontend built together | Add API endpoints as React components need them, not all upfront | — Pending |
| Keep all existing features | No capability regression during migration | — Pending |

---
*Last updated: 2026-05-11 after initialization*
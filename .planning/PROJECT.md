# Podgrab

## What This Is

Podgrab is a self-hosted podcast manager that automatically downloads podcast episodes as they become live. Built in Go with a server-side rendered web UI (Gin + Vue 2 + SQLite), it lets users subscribe to podcasts, search iTunes/PodcastIndex/gPodder, download episodes, organize with tags, and play episodes in-browser with real-time sync across tabs.

The immediate focus is stabilizing the existing Go app — upgrading outdated dependencies, fixing known bugs, modernizing error handling, and adding test coverage so scheduling and downloading work reliably. React UI modernization and REST API layer come after the base is solid and assessed.

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

- [ ] Upgrade Go to 1.24+ and all dependencies to current versions (GORM v1.26+, Gin latest, robfig/cron/v3, google/uuid, golang-jwt/jwt/v5)
- [ ] Fix known bugs (typo handler, `removeStartingSlash`, download concurrency batching, date parsing, DB init error swallowed)
- [ ] Move hardcoded PodcastIndex API credentials to environment variables
- [ ] Modernize error handling (replace `fmt.Println` with structured logging, propagate errors properly, fatal on DB init failure)
- [ ] Add thread-safe state management for WebSocket maps (replace bare maps with `sync.Map` or mutex-protected)
- [ ] Add test coverage for service and DB layers to verify fixes and prevent regressions

### Out of Scope

- React SPA / frontend modernization — deferred until after stabilization assessment
- REST API layer — deferred until after stabilization assessment
- Responsive design / dark mode / UI QoL features — deferred until after stabilization assessment
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
- Dev environment: air hot-reload on port 8080, nginx forwarding 8084→8080, LaunchAgent for boot persistence, CHECK_FREQUENCY=1 for fast iteration
- Production Docker: port 8084, CHECK_FREQUENCY=30, same volume mounts

## Constraints

- **Tech Stack**: Go backend (must stay Go), SQLite via GORM (must stay SQLite)
- **Stabilize first**: Dependency upgrades and bug fixes before any new features or UI work
- **Compatibility**: Existing data (SQLite DB, downloaded files, OPML imports) must continue to work
- **Deployment**: Must remain Docker-deployable with the same volume mount pattern (`/config`, `/assets`)
- **No feature loss**: Every current capability (search, download, tags, OPML, gPodder, backups, player sync) must be preserved
- **Dev environment**: air hot-reload workflow must continue to work throughout stabilization

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Stabilize before modernize | Scheduling/downloading must be reliable before adding React UI or API layer | — Pending |
| Tests alongside fixes | Add tests to verify bug fixes and prevent regressions | — Pending |
| React for future frontend | User chose React over Svelte/Angular for future SPA | — Pending |
| Incremental migration (future) | Keep existing features working throughout, avoid big-bang cutover | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-12 after questioning revision — stabilize-first scope*
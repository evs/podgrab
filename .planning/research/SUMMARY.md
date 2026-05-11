# Project Research Summary

**Project:** Podgrab — Self-hosted podcast manager modernization
**Domain:** Brownfield Go monolith (Go 1.15/Gin 1.7/GORM 1.20 + Vue 2 SSR) migrating to Go 1.24/Gin 1.12/GORM 1.26 + React 19 SPA
**Researched:** 2026-05-12
**Confidence:** HIGH

## Executive Summary

Podgrab is a self-hosted, download-first, single-user podcast manager built in Go with SQLite. Its niche is simplicity — lower resource overhead than Audiobookshelf (Node.js, multi-user, audiobooks) and a more mature ecosystem than PodFetch-Rust. The modernization faces two simultaneous challenges: upgrading a Go codebase across 9 major Go versions (1.15→1.24) with multiple abandoned dependencies (jwt-go, satori/uuid, gocron), and replacing a server-rendered Vue 2 UI with a React 19 SPA — all without a flag-day cutover that leaves users without a working app.

Research strongly recommends the **Strangler Fig pattern**: the Go server becomes a dual-route gateway serving both legacy SSR pages and the new React SPA simultaneously, with a new `/api/v1/*` REST API layer feeding the SPA. The most critical finding is that dependency upgrades (Go version, GORM v1→v2, abandoned libraries) must happen in isolated, single-change steps before any new code — GORM v2 has silent behavioral changes that compile cleanly but break at runtime (global update blocking, hook signature changes, tag format shifts). Equally critical is fixing the global `db.DB` singleton and WebSocket concurrency bugs before adding REST API routes, because every new endpoint amplifies the existing race conditions.

## Key Findings

### Recommended Stack

The stack is a conservative upgrade of what exists, not a ground-up rewrite. Backend stays Go+Gin+GORM+SQLite but jumps to current stable versions. Frontend switches from Vue 2 SSR to React 19 SPA with Vite 8, using TanStack Query for server state and Zustand for client state. This minimizes migration risk while delivering a modern, maintainable SPA. See `.planning/research/STACK.md` for full version details and alternatives considered.

**Core technologies:**
- **Go 1.24.x**: Stable LTS; 9-version jump from 1.15 brings structured logging, range-over-int, generics, improved routing
- **Gin v1.12.0**: Direct upgrade from 1.7.2; backward-compatible API, adds HTTP/3, BSON, better binding
- **GORM v1.26.0**: Upgrade from 1.20.2; generics API is opt-in, traditional API mostly compatible (but see pitfalls)
- **React 19.2 + Vite 8**: Current stable; no Babel needed with Oxc transformer, instant HMR
- **TanStack Query v5 + Zustand v5**: Server state + client state; eliminates Redux overkill
- **Tailwind CSS v4**: Rust-based Oxide engine, CSS-native config, `@tailwindcss/vite` plugin

**Critical upgrades (abandoned dependencies):**
- `dgrijalva/jwt-go` → `golang-jwt/jwt/v5` (security)
- `satori/go.uuid` → `google/uuid` (collision bugs)
- `jasonlvhit/gocron` → `robfig/cron/v3` (unmaintained)

### Expected Features

Podgrab's competitive position is "download-first, single-user, Go-simple." The modernization must reach parity with Audiobookshelf and PodFetch on basic QoL without bloating into a media server. See `.planning/research/FEATURES.md` for competitive analysis and full feature dependency graph.

**Must have (table stakes — ship with React migration):**
- Responsive design — without this, migration is a downgrade for mobile users
- Persistent audio player with playback speed control — every podcast app since 2018 has this
- Episode play progress tracking (persist to DB) — users expect to resume where they left off
- Played/unplayed episode indicators — visual distinction between new/listened/completed
- Dark mode — existed in original Podgrab, lost; every competitor has it
- Episode sorting (newest/oldest/alpha) and download status indicators

**Should have (competitive differentiators — build after core is solid):**
- Per-podcast auto-download rules — no competitor does per-podcast control well
- Disk space management / auto-cleanup — Podgrab's download-first nature makes this critical
- Episode timeline / "what's new" dashboard — Audiobookshelf and PodFetch have this
- Batch episode actions — select multiple → mark played, delete, download
- gPodder bi-directional sync — PodFetch has it; Podgrab has the search API already
- PWA support — 80% of native mobile benefit at 10% of the cost

**Defer (v2+):**
- Podcast 2.0 chapters/transcripts, streaming playback, HLS
- AI features, multi-user, audiobooks, social, native mobile, value4value

### Architecture Approach

The Go server becomes a **dual-route gateway** using the Strangler Fig pattern. Old SSR pages (`/`, `/add`, `/podcasts/:id/view`) coexist with the React SPA (`/app/*`) and new REST API (`/api/v1/*`). Both UIs share the same service layer and database, ensuring data consistency during migration. Pages are replaced one at a time — when the React equivalent is feature-complete, the old SSR route redirects to `/app/*`. The React SPA is embedded in the Go binary via `//go:embed`, preserving the single-binary deployment model. See `.planning/research/ARCHITECTURE.md` for component map, data flow diagrams, and Phase A-H migration sequence.

**Major components:**
1. **API Router** (`/api/v1/*`) — new REST JSON endpoints with DTO layer, proper HTTP methods, pagination
2. **SPA Fallback** (`/app/*`) — serves React SPA's `index.html`, React Router handles client-side routing
3. **Service Layer** (shared) — both old controllers and new API handlers call same `service.*` functions
4. **DTO Package** (`api/dto/`) — reshapes GORM models into clean API contracts; prevents leaking DB internals

### Critical Pitfalls

1. **GORM v1→v2 silent behavioral changes** — `BlockGlobalUpdate` default blocks deletes/updates without WHERE; hook signatures changed; model tags from snake_case ignored; `Count()` now requires `*int64`. All compile fine but break at runtime. Read release notes in full, write integration tests before upgrading, audit every model struct. (Pitfall 1)

2. **Go 1.15→1.24 breaks `ioutil` and exposes latent bugs** — `io/ioutil` removed in 1.22+; `dgrijalva/jwt-go` `+incompatible` handling changes; `go vet` flags previously-silent issues; `fmt.Println` format bugs surface. Upgrade Go FIRST in complete isolation from all other changes. (Pitfall 4)

3. **Global `db.DB` singleton + no WAL mode** — concurrent API goroutines will escalate SQLite "database is locked" errors; can't write tests without DI. Inject repository struct, enable WAL mode, set `busy_timeout` BEFORE adding REST API routes. (Pitfall 5)

4. **Dual-frontend state divergence** — Vue 2 and React UIs showing different data for same operations because template rendering and API serialization use different code paths. Define DTOs that both use; make both go through service layer; integration tests for data parity. (Pitfall 3)

5. **SQL injection via sort params in API** — `db.GetAllPodcasts` passes user `sorting` string directly to `DB.Order()`. Safe today only because template pages use hardcoded values. Whitelist allowed sort columns before any API endpoint accepts user-supplied sort/filter. (Pitfall 10)

## Implications for Roadmap

Based on combined research, the phase structure follows the dependency chain: backend stability must precede API layer, which must precede React UI, which must precede feature additions. Pitfalls force a specific ordering — Go upgrade isolated first, then GORM + dependency upgrades one-at-a-time, then infrastructure fixes (DI, WAL, mutexes), then API + SPA scaffold, then page-by-page migration.

### Phase 1: Backend Foundation
**Rationale:** Every subsequent phase depends on a stable, current backend. Go version upgrade must be isolated (Pitfall 4). GORM upgrade has silent breaks that must be caught before any new DB operations (Pitfall 1). Abandoned dependencies are security/compatibility risks. WAL mode and repository injection prevent the concurrency crises that new API routes would trigger (Pitfall 5).
**Delivers:** Go 1.24, current GORM, all deps upgraded, WAL mode enabled, repository struct injected, `ioutil` replaced, `fmt.Println` → zap, WebSocket mutex fixes
**Addresses:** Dependency upgrades, security fixes, concurrency safety
**Avoids:** Pitfalls 1, 4, 5, 11, 12, 15

### Phase 2: REST API + SPA Infrastructure
**Rationale:** API layer and SPA scaffold must exist before any React component can be built. DTO layer prevents dual-frontend data divergence (Pitfall 3). `go:embed` and multi-stage Docker solve build/deploy collision (Pitfall 7). API route namespacing + Gin catch-all prevents routing conflicts (Pitfall 14). Sort param whitelisting prevents SQL injection (Pitfall 10).
**Delivers:** `/api/v1/*` route group, DTO package, API middleware (auth, CORS for dev), SPA scaffold (Vite+React+Router+TanStack Query+Zustand+Tailwind), `go:embed` integration, multi-stage Dockerfile, SPA fallback route
**Uses:** Gin route groups, React 19, Vite 8, TanStack Query v5, Zustand v5, Tailwind CSS v4
**Implements:** API Router, SPA Fallback, DTO Package components
**Avoids:** Pitfalls 3, 7, 9, 10, 14

### Phase 3: Core Views — Podcast List & Detail
**Rationale:** First concrete SPA pages validate the entire stack end-to-end. Podcast list is the landing page — gets the most traffic. Podcast detail with episode list is the most data-heavy view, testing pagination and DTO shaping.
**Delivers:** `/app/` podcast list, `/app/podcasts/:id` detail with episodes, `/app/add` search+subscribe, API endpoints for podcasts/episodes/search
**Addresses:** Responsive design (table stakes), episode sorting, download status indicators
**Implements:** Strangler Fig Phase B, C, D from architecture

### Phase 4: Audio Player & Progress
**Rationale:** Player is the most complex component — HTML5 Audio + WebSocket sync + persistent state. Must come after team is comfortable with SPA patterns. WebSocket protocol must support both old and new clients simultaneously (Pitfall 8). Play progress tracking enables played/unplayed indicators (feature dependency chain).
**Delivers:** Persistent audio player bar, playback speed control, play progress persistence to DB, played/unplayed indicators, keyboard shortcuts
**Addresses:** Table stakes: persistent player, playback speed, progress tracking, played indicators, keyboard shortcuts
**Avoids:** Pitfall 2 (WebSocket concurrency already fixed in Phase 1), Pitfall 8
**Implements:** Strangler Fig Phase F

### Phase 5: Settings, Tags, OPML & Remaining Views
**Rationale:** Less complex views that complete the SPA feature set. Tags use existing service-layer bypasses that must be routed through service first (Pattern 3). OPML has uncapped goroutines that must be bounded (Pitfall 13).
**Delivers:** `/app/tags`, `/app/settings`, `/app/backups`, OPML import/export, remaining API endpoints, batch episode actions
**Addresses:** Episode search/filter, RSS re-scan on demand, batch actions
**Avoids:** Pitfall 13 (OPML goroutine limits)
**Implements:** Strangler Fig Phase E, G

### Phase 6: QoL Features & Differentiators
**Rationale:** Core SPA is complete. Now add the features that elevate Podgrab above "it works" to "it's pleasant." Dark mode, timeline dashboard, per-podcast auto-download rules, disk management — these depend on the foundation being solid.
**Delivers:** Dark mode, episode timeline/"what's new" dashboard, per-podcast auto-download rules, disk space management/auto-cleanup, subscription health/feed errors, episode description/notes rendering, PWA support
**Addresses:** Remaining table stakes (dark mode, episode notes) and Priority 2-3 differentiators

### Phase 7: SSR Cleanup
**Rationale:** Only after ALL pages have React equivalents and are tested. Removing legacy code is satisfying but irreversible — one redirect mistake and users hit 404s. Must be the last phase.
**Delivers:** Remove Go templates, remove webassets, remove template FuncMap, make `/app/*` the default route, remove legacy page controllers
**Implements:** Strangler Fig Phase H

### Phase Ordering Rationale

- **Backend before frontend:** Dependency upgrades and infrastructure fixes (Phase 1) must complete before any new code depends on them. GORM silent breaks (Pitfall 1) and Go `ioutil` removal (Pitfall 4) are compilation or runtime blockers.
- **Infrastructure before UI:** API layer + SPA scaffold (Phase 2) must exist before React components (Phases 3-5) can be built. DTOs and service-layer contracts prevent the dual-frontend divergence trap (Pitfall 3).
- **Pages in dependency order:** Podcast list → detail → player → settings follows the Strangler Fig migration sequence from architecture research, which is ordered by data dependency complexity.
- **Features after foundation:** Dark mode, timeline, disk management (Phase 6) depend on the SPA and API being stable. Adding them earlier would complicate the migration.
- **Cleanup last:** Removing legacy SSR code (Phase 7) is irreversible and blocks rollback if done too early.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 1 (Backend Foundation):** GORM v1→v2 migration requires careful audit of every model and query — read release notes thoroughly, test each model's behavior individually
- **Phase 4 (Audio Player):** WebSocket protocol documentation needed before React client; audio library choice (HTML5 Audio vs Howler.js) needs spike; persistent progress DB schema design

Phases with standard patterns (skip research-phase):
- **Phase 2 (Infrastructure):** Well-documented patterns — Gin route groups, Vite+React scaffold, `go:embed`, multi-stage Docker
- **Phase 3 (Core Views):** Standard REST CRUD + React list/detail pages — well-established patterns
- **Phase 5 (Remaining Views):** Follows same patterns as Phase 3
- **Phase 7 (Cleanup):** Mechanical removal of legacy code

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified against official sources (Go proxy, npm registry, GitHub releases) on 2026-05-12 |
| Features | HIGH | Competitive analysis verified against Audiobookshelf docs, PodFetch docs, Podverse README, original Podgrab issues |
| Architecture | HIGH | Strangler Fig is proven pattern; codebase analysis is first-hand; Gin/Vue coexistence validated by reading existing routes |
| Pitfalls | HIGH | GORM v2 breaking changes from official docs; Go release notes official; concurrency issues verified in codebase |

**Overall confidence:** HIGH

### Gaps to Address

- **WebSocket protocol specification:** No documentation exists for the current message format. Must be reverse-engineered from `controllers/websockets.go` and documented before Phase 4. A protocol doc should be a Phase 1 or Phase 2 deliverable.
- **GORM model audit completeness:** The full list of model structs with snake_case tags, hook signatures, and `Count()` call signatures needs a systematic audit. Recommended as a Phase 1 planning task (enumerate all models before upgrading).
- **shadcn/ui + Tailwind v4 compatibility:** shadcn/ui v4.7 supports Tailwind v4 but the integration details (CSS variables, theme configuration) may need a spike during Phase 2 to verify the setup works cleanly.
- **Existing test coverage:** The codebase appears to have minimal test coverage. Phase 1 should establish a test harness even if tests are sparse — the integration test infrastructure needs to exist before GORM migration.

## Sources

### Primary (HIGH confidence)
- Go official release notes — go.dev/doc/devel/release (Go 1.24, 1.25, 1.26 version verification)
- Gin GitHub releases — github.com/gin-gonic/gin/releases (v1.12.0)
- GORM v2 Release Note — gorm.io/docs/v2_release_note.html (breaking changes)
- golang-jwt Migration Guide — github.com/golang-jwt/jwt/MIGRATION_GUIDE.md
- Martin Fowler — Strangler Fig Application — martinfowler.com/bliki/StranglerFigApplication.html
- Go `//go:embed` documentation — pkg.go.dev/embed
- npm registry — React 19.2, Vite 8, TanStack Query v5, Zustand v5, Tailwind CSS v4 (version verification)
- Audiobookshelf official docs — audiobookshelf.org/docs
- PodFetch docs — podfetch-docs.samtv.fyi
- Podgrab codebase analysis — .planning/codebase/ (first-hand inspection)

### Secondary (MEDIUM confidence)
- Podverse GitHub README — cloud-hosted model differs from self-hosted, but feature patterns are valid
- React.dev docs — SSR hydration focus differs from server-template-to-SPA migration, but component patterns are standard
- shadcn/ui + Tailwind v4 — integration exists but needs verification with the specific versions chosen

### Tertiary (LOW confidence)
- Community discussions on gorilla/websocket replacement —Coder/websocket vs Gorilla choice may shift with new releases; research recommends staying with Gorilla v1.5.3 (now actively maintained again) per STACK.md

---
*Research completed: 2026-05-12*
*Ready for roadmap: yes*
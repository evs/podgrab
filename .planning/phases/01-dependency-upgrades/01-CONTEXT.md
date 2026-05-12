# Phase 1: Dependency Upgrades - Context

**Gathered:** 2026-05-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Upgrade Go to 1.24+, replace all abandoned dependencies with maintained alternatives, move hardcoded API credentials to environment variables, update Docker configuration. The app must compile, run, and pass all existing functionality after each incremental step — the download-and-organize loop must never break.

</domain>

<decisions>
## Implementation Decisions

### Dependency Version Targets

- **D-01:** Go 1.24.x target (not 1.25/1.26 — more battle-tested, stable choice)
- **D-02:** GORM v1.26.x (not v1.31 — conservative, proven, traditional API still works)
- **D-03:** Gin upgraded to latest stable (v1.12+)
- **D-04:** `satori/go.uuid` → `google/uuid` (API-compatible drop-in, maintained)
- **D-05:** `dgrijalva/jwt-go` → **removed entirely** (unused in codebase, security liability)
- **D-06:** `jasonlvhit/gocron` → `robfig/cron/v3` (standard Go cron with expression syntax)
- **D-07:** `gorilla/websocket` → `coder/websocket` (maintained fork by Coder, API-compatible)
- **D-08:** All other dependencies upgraded to current via systematic `go get -u`

### Upgrade Strategy

- **D-09:** Incremental upgrade sequence — Go first, then swap abandoned libs one at a time, then GORM last. Each step produces a working binary. GORM v1→v2 has silent runtime behavior changes so it must come last when we can test against a stable app.

### Environment & Configuration

- **D-10:** PodcastIndex API credentials (`PODCASTINDEX_KEY`, `PODCASTINDEX_SECRET`) moved to `.env` file pattern, consistent with existing `PASSWORD`, `CONFIG`, `DATA`, `CHECK_FREQUENCY` env vars
- **D-11:** Dockerfile updated with Go 1.24+ base image in multi-stage build, keeping same `/config` and `/assets` volume mount pattern

### Code Changes

- **D-12:** Deprecated `io/ioutil` calls replaced with `io.ReadAll`, `os.WriteFile` etc.
- **D-13:** GORM v2 migration notes from research to be followed carefully (soft delete, hook signatures, method chain safety, tag changes)

### the agent's Discretion

- Specific Go 1.24.x minor version (1.24.1, 1.24.2, etc.) — use latest stable patch
- Specific GORM v1.26.x minor version — use latest stable patch
- Order of abandoned lib swaps within the incremental sequence
- Whether to update `go.mod` Go directive before or alongside dependency bumps (agent decides based on compatibility)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Dependency Research
- `.planning/research/STACK.md` — Specific version recommendations, rationale, and anti-recommendations for all dependency upgrades
- `.planning/research/PITFALLS.md` — GORM v1→v2 behavioral breaking changes, Go version upgrade issues, and phase-specific warnings

### Known Concerns
- `.planning/codebase/CONCERNS.md` — Tech debt inventory (hardcoded credentials, deprecated ioutil, outdated Go version), security issues, and known bugs
- `.planning/codebase/STACK.md` — Current dependency versions and configuration patterns

### Project Context
- `.planning/PROJECT.md` — Core value, constraints, and requirements
- `.planning/REQUIREMENTS.md` — DEPS-01 through DEPS-10 requirements
- `.planning/ROADMAP.md` — Phase 1 success criteria

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `service/itunesService.go:42-44` — Location of hardcoded PodcastIndex credentials that need env var extraction
- `godotenv/autoload` import pattern — Already used project-wide for `.env` loading; new env vars follow same pattern
- `main.go:220` — Cron scheduling with `gocron`; needs rewrite to `robfig/cron/v3` expression syntax
- `db/db.go:15` — Global `db.DB` singleton; remains as-is in this phase (global refactoring is Phase 2+)
- `controllers/websockets.go:22-25` — WebSocket maps using `gorilla/websocket`; need `coder/websocket` API adaptation

### Established Patterns
- Environment variable pattern: `os.Getenv("VAR_NAME")` used for `PASSWORD`, `CONFIG`, `DATA`, `CHECK_FREQUENCY` in `main.go` — PodcastIndex env vars follow this pattern exactly
- Docker multi-stage build pattern: `Dockerfile` uses `golang:1.15-alpine` builder → `alpine:latest` runtime — update builder image only
- `.env` file: Already loaded via `_ "github.com/joho/godotenv/autoload"` — new vars just need adding to `.env`

### Integration Points
- `go.mod` — Central file for all version changes; every dependency swap touches this
- `main.go` — Entry point for cron scheduling rewrite (`intiCron` function), env var loading, template functions
- `Dockerfile` / `docker-compose.yml` — Go version in builder stage, environment variable additions for new secrets

</code_context>

<specifics>
## Specific Ideas

No specific requirements — incremental dependency upgrade is a well-understood process. Follow the research recommendations and GORM migration notes carefully.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.
</deferred>

---

*Phase: 1-Dependency Upgrades*
*Context gathered: 2026-05-12*
# Phase 1: Dependency Upgrades - Research

**Phase:** 1
**Gathered:** 2026-05-12
**Status:** Referenced from project-level research

## Source Documents

This phase's research is drawn from the project-level research documents:

- `.planning/research/STACK.md` — Dependency version targets, upgrade paths, migration effort estimates, installation commands
- `.planning/research/PITFALLS.md` — GORM v1→v2 silent breaks, Go version upgrade issues, WebSocket replacement risks, ioutil deprecation
- `.planning/research/FEATURES.md` — Feature inventory for validation after upgrades
- `.planning/research/ARCHITECTURE.md` — Component map, data flows, entry points for change impact analysis
- `.planning/research/SUMMARY.md` — Research synthesis

## Key Findings for Phase 1

### Go 1.15 → 1.24 Upgrade (DEPS-01)

- `io/ioutil` is deprecated since Go 1.16, removed in Go 1.22+ — must replace before Go 1.24 compiles
- `go.mod` says `go 1.15` (pre-module mode) — `go mod tidy` behavior changes significantly
- `dgrijalva/jwt-go v3.2.0+incompatible` — `+incompatible` flag handled differently by newer Go toolchains
- `golang.org/x/crypto` and `golang.org/x/net` have known CVEs at current versions
- New `go vet` checks will flag issues (e.g., `fmt.Println` with format verbs)

**Strategy:** Upgrade Go FIRST before any other dependency changes. Isolate the blast radius.

### GORM v1.20 → v1.26 (DEPS-02)

Critical silent breaking changes:
- `BlockGlobalUpdate` is now default — `Delete`/`Update` without `WHERE` silently do nothing
- Model tags changed from `snake_case` to `camelCase` — `auto_increment`, `unique_index` silently ignored
- `db.CreateTable()` / `db.DropTable()` no longer exist → use `db.Migrator()`
- `TableName()` return value now cached — dynamic table names break
- `Soft Delete` requires `gorm.DeletedAt` type, not just a field named `DeletedAt`
- Hook signatures changed from `func(u *User) error` to `func(tx *gorm.DB) error`
- `Count()` only accepts `*int64` — existing `int` or `int32` calls fail silently
- Method chain safety: reusing `*gorm.DB` after `.Where()` in goroutines causes data races

**Strategy:** Upgrade GORM LAST after all other upgrades are stable. Audit every model struct, hook, query before the upgrade.

### Abandoned Library Replacements

#### satori/go.uuid → google/uuid (DEPS-03)
- Near drop-in replacement: `uuid.New()` returns `[16]byte`, not string
- Existing UUIDs stored as strings in SQLite remain valid
- Do in ONE commit, avoid importing both libraries simultaneously
- Current usage: `db/base.go` BeforeCreate hook generates UUIDs for all models

#### dgrijalva/jwt-go → REMOVED (DEPS-04)
- **Confirmed unused** — no imports found in codebase
- Remove from go.mod entirely, no replacement needed
- Context.md decision D-05: "removed entirely (unused in codebase, security liability)"

#### jasonlvhit/gocron → robfig/cron/v3 (DEPS-05)
- Completely different API — chained scheduler becomes cron expression syntax
- `main.go:220` has the cron setup that needs rewrite
- Current scheduler: `s.Every(30).Minutes().Do(task)` → becomes `c.AddFunc("*/30 * * * *", task)`
- Need to handle the `initCron` function with `robfig/cron/v3` syntax

#### gorilla/websocket → coder/websocket (DEPS-07, per CONTEXT.md D-07)
- CONTEXT.md decision: `coder/websocket` (maintained fork by Coder)
- STACK.md research recommended staying with `gorilla/websocket` v1.5.3
- **CONFLICT:** CONTEXT.md chose `coder/websocket`, STACK.md suggested staying with Gorilla v1.5.3
- Per CONTEXT.md D-07 decision, use `coder/websocket`
- API differences: `coder/websocket` uses different connection upgrade pattern
- `controllers/websockets.go:22-25` — maps need API adaptation

### Gin 1.7.2 → 1.12+ (DEPS-06)
- Mostly backward compatible
- Minimum Go 1.21 requirement (already satisfied by Go 1.24)
- New features: BSON protocol, protobuf content negotiation
- API is backward-compatible

### Deprecated ioutil Replacement (DEPS-08)
- `ioutil.ReadAll` → `io.ReadAll`
- `ioutil.WriteFile` → `os.WriteFile`
- `ioutil.ReadFile` → `os.ReadFile`
- `ioutil.TempDir` → `os.MkdirTemp`
- `ioutil.TempFile` → `os.CreateTemp`
- Files affected: `service/podcastService.go`, `service/fileService.go`

### Hardcoded PodcastIndex Credentials (DEPS-09)
- Location: `service/itunesService.go` lines 42-44
- Move to environment variables: `PODCASTINDEX_KEY`, `PODCASTINDEX_SECRET`
- Already consistent with existing pattern: `PASSWORD`, `CONFIG`, `DATA`, `CHECK_FREQUENCY` all use `os.Getenv()`
- `.env` file already loaded via `_ "github.com/joho/godotenv/autoload"`

### Docker Updates (DEPS-10)
- `Dockerfile` uses `golang:1.15-alpine` builder → update to `golang:1.24-alpine` or `golang:1.24`
- `docker-compose.yml` — add `PODCASTINDEX_KEY` and `PODCASTINDEX_SECRET` environment variables
- Keep same `/config` and `/assets` volume mount pattern
- Keep multi-stage build pattern

### Recommended Upgrade Sequence

Per CONTEXT.md D-09 (incremental sequence):
1. Go version upgrade (`go.mod` directive + `ioutil` replacement first)
2. Swap `dgrijalva/jwt-go` → remove (unused)
3. Swap `satori/go.uuid` → `google/uuid`
4. Swap `jasonlvhit/gocron` → `robfig/cron/v3`
5. Swap `gorilla/websocket` → `coder/websocket`
6. Upgrade all remaining dependencies (`gin`, `zap`, `godotenv`, `x/crypto`, `x/net`, etc.)
7. Upgrade GORM (LAST — most risky, requires careful testing)
8. Update Docker configuration

Each step must produce a working binary before proceeding to the next.

### Special Concerns

- **`db/base.go` BeforeCreate hook:** Uses `satori/go.uuid` for UUID generation. When swapping to `google/uuid`, the `BeforeCreate` hook signature must be verified — GORM v1 uses `func(u *User) error` but v2 uses `func(tx *gorm.DB) error`. This is a double-hit: the UUID library AND the hook signature change.
- **Global `db.DB` singleton:** Remains as-is in this phase (Phase 2+ concern). But GORM upgrade must not break the global access pattern.
- **Cron scheduler rewrite:** `initCron` function in `main.go` must be completely rewritten for `robfig/cron/v3`. The current gocron syntax is fundamentally different.
- **WebSocket upgrade:** `coder/websocket` has a different API than `gorilla/websocket`. The upgrade handler, message types, and connection lifecycle all need adaptation.

---

*Research referenced: 2026-05-12*
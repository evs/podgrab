<!-- GSD:project-start source:PROJECT.md -->
## Project

**Podgrab**

Podgrab is a self-hosted podcast manager that automatically downloads podcast episodes as they become live. Built in Go with a server-side rendered web UI (Gin + Vue 2 + SQLite), it lets users subscribe to podcasts, search iTunes/PodcastIndex/gPodder, download episodes, organize with tags, and play episodes in-browser with real-time sync across tabs.

The immediate focus is stabilizing the existing Go app — upgrading outdated dependencies, fixing known bugs, modernizing error handling, and adding test coverage so scheduling and downloading work reliably. React UI modernization and REST API layer come after the base is solid and assessed.

**Core Value:** Podcast episodes are automatically downloaded and available — that download-and-organize loop must never break.

### Constraints

- **Tech Stack**: Go backend (must stay Go), SQLite via GORM (must stay SQLite)
- **Stabilize first**: Dependency upgrades and bug fixes before any new features or UI work
- **Compatibility**: Existing data (SQLite DB, downloaded files, OPML imports) must continue to work
- **Deployment**: Must remain Docker-deployable with the same volume mount pattern (`/config`, `/assets`)
- **No feature loss**: Every current capability (search, download, tags, OPML, gPodder, backups, player sync) must be preserved
- **Dev environment**: air hot-reload workflow must continue to work throughout stabilization
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.15 — Backend server, all business logic, controllers, models, services
- HTML/JavaScript/CSS — Frontend templates served via Go's `html/template` (server-side rendered)
## Runtime
- Go runtime (compiled binary)
- Docker container (Alpine Linux base image)
- `GIN_MODE=release` for production
- Go Modules — `go.mod` present
- Lockfile: `go.sum` present
## Frameworks
- Gin v1.7.2 — HTTP web framework (`github.com/gin-gonic/gin`)
- Gin Contrib Location v0.0.2 — URL/location helper (`github.com/gin-contrib/location`)
- Gorilla WebSocket v1.4.2 — WebSocket support for real-time player communication
- GORM v1.20.2 — ORM for SQLite (`gorm.io/gorm`)
- GORM SQLite Driver v1.1.3 — SQLite dialect (`gorm.io/driver/sqlite`)
- Not detected — no test files found in the repository
- Docker — multi-stage build (Go builder → Alpine runtime)
- Docker Compose v2.1 — container orchestration
## Key Dependencies
- `github.com/TheHippo/podcastindex` v1.0.0 — Podcast Index API client for podcast search
- `github.com/antchfx/xmlquery` v1.3.3 — XML DOM querying (RSS feed image extraction)
- `github.com/dgrijalva/jwt-go` v3.2.0+incompatible — JWT token support (unused or future auth)
- `golang.org/x/crypto` — Cryptographic functions ( bcrypt password hashing)
- `go.uber.org/zap` v1.16.0 — Structured logging
- `github.com/joho/godotenv` v1.3.0 — `.env` file loading (autoloaded via `_ "github.com/joho/godotenv/autoload"`)
- `github.com/jasonlvhit/gocron` v0.0.1 — Job scheduling (episode refresh, file checks, backups)
- `github.com/satori/go.uuid` v1.2.0 — UUID generation for model primary keys
- `github.com/gobeam/stringy` v0.0.0-20200717095810-8a3637503f62 — String manipulation (kebab-case for filenames)
- `github.com/grokify/html-strip-tags-go` v0.0.0-20200923094847-079d207a09f1 — HTML tag stripping
- `github.com/microcosm-cc/bluemonday` v1.0.15 — HTML sanitization policy
- `github.com/chris-ramon/douceur` v0.2.0 — CSS sanitization (indirect dependency)
## Configuration
- `.env` file via `godotenv/autoload` — contains secrets/config (never read in code)
- `CONFIG` env var — Path to configuration directory (database, backups)
- `DATA` env var — Path to assets/data directory (downloaded podcast files)
- `PASSWORD` env var — Optional HTTP Basic Auth credentials (user: "podgrab")
- `CHECK_FREQUENCY` env var — Cron check interval in minutes (default: 30)
- `PUID` / `PGID` env var — File ownership UID/GID for downloaded files
- `GIN_MODE` env var — Gin framework mode (`release` for production)
- `Dockerfile` — Multi-stage Go build, Alpine runtime
- `docker-compose.yml` — Service orchestration with volume mounts
## Platform Requirements
- Go 1.15+ (specified in `go.mod`)
- SQLite3 (via CGO — `gorm.io/driver/sqlite` requires `gcc` and `CGO_ENABLED=1`)
- Docker (Alpine-based image `akhilrex/podgrab`)
- Persistent volumes: `/config` (database, backups), `/assets` (downloaded media)
- Port 8080 (default, hardcoded in `main.go` via `r.Run()`)
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Go standard: lowercase single-word package names
- Packages: `db`, `model`, `service`, `controllers`, `sanitize` (under `internal/`)
- Module: `github.com/akhilrex/podgrab`
- camelCase for multi-word files: `podcastService.go`, `fileService.go`, `itunesService.go`, `gpodderService.go`, `naturaltime.go`, `podcastModels.go`, `rssModels.go`, `queryModels.go`, `gpodderModels.go`, `podcastModels.go`, `dbfunctions.go`
- Single-word files for core concepts: `db.go`, `base.go`, `errors.go`
- PascalCase for exported functions: `GetAllPodcasts`, `AddPodcast`, `FetchURL`, `DeleteFile`
- camelCase for unexported functions: `makeQuery`, `getSortOrder`, `createFolder`, `downloadImageLocally`, `deleteOldBackup`, `changeOwnership`, `getFileName`, `cleanFileName`, `checkError`
- Constructor-style not used — no `New*` factory functions (except `podcastindex.NewClient` from external lib)
- Verbose descriptive names: `SetPodcastItemAsQueuedForDownload`, `SetPodcastItemAsDownloaded`, `GetAllPodcastItemsByPodcastId`, `GetPaginatedPodcastItemsNew`
- camelCase: `podcastItem`, `searchByIdQuery`, `latestDate`, `podcastListQuery`
- Short names in loops: `obj`, `v`, `item`, `pod`
- Map entries: `countMap`, `sizeMap`, `keyMap`
- Acronyms preserved: `URL`, `GUID`, `ID` — but inconsistently: `url`, `Id` (not `ID`)
- PascalCase structs: `Podcast`, `PodcastItem`, `Setting`, `Tag`, `JobLock`
- Custom error types with struct: `PodcastAlreadyExistsError`, `TagAlreadyExistsError`
- Enums via `iota`: `DownloadStatus` (`NotDownloaded`, `Downloading`, `Downloaded`, `Deleted`)
- String enums via typed constants: `EpisodeSort` (`RELEASE_ASC`, `RELEASE_DESC`, `DURATION_ASC`, `DURATION_DESC`)
## Code Style
- No linter or formatter config files detected (`.eslintrc`, `.prettierrc`, etc.)
- Go standard `gofmt` formatting assumed
- Mixed comment styles: some `//` with text, many `//` blank comments left as placeholders
- No linting configuration detected
- No CI linting step in `.github/workflows/hub.yml` — only Docker build/push
- `fmt.Println` used extensively for logging (62+ occurrences) instead of structured logger
- Inconsistent error handling: some `fmt.Println(err)`, some `log.Println(err)`, some `Logger.Errorw(...)`
- One `panic(err)` in `service/fileService.go:407` (`checkError` function)
- One `log.Fatal` in `service/itunesService.go:52` — crashes the process on PodcastIndex search error
- Commented-out `fmt.Println` statements left in production code throughout
- Typo in function name: `DeletePodcasDeleteOnlyPodcasttEpisodesById` (line 162, `controllers/podcast.go`)
- Typo in function name: `NatualTime` (should be `NaturalTime`, `service/naturaltime.go:9`)
- Global mutable state: `db.DB` (package-level `*gorm.DB`), `var Logger`, `var activePlayers`, `var allConnections`, `var broadcast`
## Import Organization
- `stringy "github.com/gobeam/stringy"` — aliased import
- `strip "github.com/grokify/html-strip-tags-go"` — aliased import
- `uuid "github.com/satori/go.uuid"` — aliased import
- No explicit import grouping beyond what `goimports` would produce
## Error Handling
- `error` return values propagated from DB layer: `return result.Error` (`db/dbfunctions.go`)
- Custom error types implement `Error()` stringer: `model/errors.go`
- Type assertion for custom errors in controllers: `if v, ok := err.(*model.PodcastAlreadyExistsError); ok`
- `errors.Is(err, gorm.ErrRecordNotFound)` for GORM-specific checks
- Many errors silently consumed: `fmt.Println(err)` without returning or handling
- `ShouldBind*` / `ShouldBindQuery` / `ShouldBindUri` used for request validation — errors often result in `gin.H{"error": "Invalid request"}`
- Errors printed but not returned: `fmt.Println(err)` after `db.GetPodcastById` (lines 92, 182, 195, 231 in `controllers/podcast.go`)
- Error swallowed in `makeQuery`: `body, _ := makeQuery(url)` in `service/gpodderService.go:19` and `service/itunesService.go:25`
- Mixed `fmt.Println` vs `log.Println` vs `Logger.Errorw` for same conceptual purpose
## Logging
- Structured logging via `Logger.Errorw("message", err)` in select places (`service/podcastService.go`)
- Most of the codebase uses `fmt.Println` / `fmt.Printf` for debug output
- `log.Println` / `log.Print` used sparsely in `main.go` and `db/db.go`
- No log levels configured — `zap.NewProduction()` in init()
- Errors from HTTP requests: mixed `fmt.Println`, `log.Println`
- Debug info: `fmt.Println` scattered throughout cron jobs and websocket handlers
- Structured errors: `Logger.Errorw` only in `service/podcastService.go`
## Comments
- Structural comments: `//PodcastData is`, `//PodcastItem is`, `//Base is`, `//BeforeCreate` — stub comments that add no value
- Many `//` blank line comments left behind (e.g., `//fmt.Println(...)` — commented-out debug)
- No doc comments on exported functions — Go convention would expect them
- Inline explanation for date parsing chain in `service/podcastService.go` (RFC1123 format variations)
- Not applicable — Go project
- Go doc comments are mostly absent on exported symbols
## Function Design
- `AddPodcastItems` in `service/podcastService.go` ~110 lines
- `PlayerPage` in `controllers/pages.go` ~50 lines
- Template function map in `main.go` spans ~100 lines inline
- `UpdateSettings` takes 12 parameters (`service/podcastService.go:766`)
- `Download` takes 4 parameters (`service/fileService.go:25`)
- Pointer receivers used for mutation, value types for queries
- DB functions: `(result, error)` tuple convention
- Service functions: sometimes return pointer, sometimes value — inconsistent
- `GetOrCreateSetting` returns `*Setting` with no error (panics on failure silently handled via GORM)
## Module Design
- All public symbols in a package are accessible (Go visibility rules)
- No `internal/` boundary enforcement except `internal/sanitize`
- `db` package leaks `DB` as exported global variable — accessible from all packages
- No explicit barrel/index files
- Each package exposes its types through individual files
- `db/base.go` defines shared `Base` struct used across all DB models
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## System Overview
```text
```
## Component Responsibilities
| Component | Responsibility | File |
|-----------|----------------|------|
| Router | HTTP routing, middleware, template engine setup | `main.go` |
| Page Controllers | Server-side rendered HTML pages | `controllers/pages.go` |
| API Controllers | REST JSON/XML endpoints for podcasts, episodes, tags | `controllers/podcast.go` |
| WebSocket Handler | Real-time player state sync across clients | `controllers/websockets.go` |
| Podcast Service | Core podcast/episode business logic — add, refresh, download, delete | `service/podcastService.go` |
| File Service | File download, backup creation, image handling, NFO generation | `service/fileService.go` |
| Search Service | iTunes and PodcastIndex podcast search | `service/itunesService.go` |
| gPodder Service | gPodder.net API integration (search, tags, top lists) | `service/gpodderService.go` |
| Natural Time | Human-readable relative time formatting | `service/naturaltime.go` |
| DB Layer | GORM models, CRUD functions, migrations, job locking | `db/` |
| Models | Data transfer objects for RSS, OPML, iTunes, queries, errors | `model/` |
| Sanitize | File name sanitization (vendored) | `internal/sanitize/sanitize.go` |
## Pattern Overview
- No separation between "handler" and "controller" — controller functions directly call service functions
- Service layer is flat (no interfaces except `SearchService` for search providers)
- Global mutable state: `db.DB` is a package-level `*gorm.DB` singleton initialized in `main.go`
- WebSocket state held in package-level maps (`activePlayers`, `allConnections`, `broadcast`)
- Cron-based background tasks scheduled in `main.go` via `gocron`
- HTML templates in `client/` mixed with inline JS/Vue.js
- Static assets in `webassets/` (vendored JS/CSS libraries, no build system)
## Layers
- Purpose: Handle HTTP requests, render templates, manage WebSocket connections
- Location: `controllers/`, `client/`
- Contains: Gin handler functions, Go HTML templates, inline JavaScript
- Depends on: `service/`, `db/`, `model/`
- Used by: Gin router in `main.go`
- Purpose: Business logic — podcast management, file I/O, search, OPML import/export
- Location: `service/`
- Contains: Pure functions and service structs (only `SearchService` interface + `ItunesService`/`PodcastIndexService` implementations)
- Depends on: `db/`, `model/`, `internal/sanitize`
- Used by: Controllers, cron jobs
- Purpose: Data persistence, database queries, migrations
- Location: `db/`
- Contains: GORM model definitions, query functions, migration runner, job locking
- Depends on: GORM, SQLite driver
- Used by: Service layer, some controllers directly
- Purpose: DTOs for external data formats (RSS, OPML, iTunes) and query parameters
- Location: `model/`
- Contains: XML/JSON deserialization structs, query/filter models, custom error types
- Depends on: Nothing (pure data structures)
- Used by: All other layers
## Data Flow
### Primary Request Path (Add Podcast)
### Episode Download Flow
### WebSocket Player Sync
- Database: SQLite via GORM (global `db.DB` variable)
- In-memory: WebSocket connection maps (`activePlayers`, `allConnections`) — not thread-safe for concurrent map writes
- Settings: `db.GetOrCreateSetting()` called per request in middleware — singleton row in SQLite
## Key Abstractions
- Purpose: Abstract podcast search across providers
- Examples: `service/itunesService.go` (`ItunesService`, `PodcastIndexService`)
- Pattern: Strategy pattern via interface `SearchService` with `Query(q string) []*model.CommonSearchResultModel`
- Purpose: All entities inherit UUID primary key + timestamps via `db.Base`
- Examples: `db.Podcast`, `db.PodcastItem`, `db.Setting`, `db.Tag`, `db.JobLock`
- Pattern: `db.Base` embeds `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`; `BeforeCreate` hook generates UUID
## Entry Points
- Location: `main.go:21`
- Triggers: Process start
- Responsibilities: DB init, Gin router setup, template loading, route registration, middleware, cron scheduling
- Location: `main.go:220`
- Triggers: On process start (background goroutine)
- Responsibilities: Periodic episode refresh, missing file check, image downloads, backup creation
- Location: `controllers/websockets.go:65`
- Triggers: WebSocket connections on `/ws`
- Responsibilities: Player state synchronization, enqueue commands
## Architectural Constraints
- **Threading:** Single-process, multi-goroutine. Download operations use `sync.WaitGroup` with concurrency limit. Cron runs in background goroutine. WebSocket handler runs in separate goroutine.
- **Global state:** `db.DB` (`db/db.go:15`), `activePlayers` / `allConnections` maps (`controllers/websockets.go:22-23`), `broadcast` channel (`controllers/websockets.go:25`). No mutex protection on WebSocket maps.
- **Circular imports:** controllers → service → db, but no reverse — one-directional dependency. controllers also imports db directly (bypassing service layer in some handlers).
- **No dependency injection:** Services are instantiated directly (`new(service.ItunesService)`) or accessed as package-level functions. No DI container.
- **SQLite-only:** Hardcoded to SQLite via GORM — no database abstraction for swapping engines.
- **Server-side rendering only:** HTML templates with embedded Vue.js for interactivity. No API/backend separation.
## Anti-Patterns
### Controllers Bypass Service Layer
### Global Mutable State Without Thread Safety
### Hardcoded API Keys
### Duplicate Controller Function
## Error Handling
- Service functions return `error` but callers frequently ignore them (e.g., `controllers/podcast.go:196`: `go service.RefreshEpisodes()` ignores return)
- `fmt.Println` used for error logging in many places (e.g., `controllers/podcast.go:92-93`, `service/podcastService.go:276`)
- `zap` logger initialized in `service/podcastService.go:24-30` but rarely used
- Panic-on-error in `service/fileService.go:405-408` (`checkError` function panics)
- No structured error types beyond `model.PodcastAlreadyExistsError` and `model.TagAlreadyExistsError`
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->

<!-- refreshed: 2026-05-11 -->
# Architecture

**Analysis Date:** 2026-05-11

## System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                      HTTP / WebSocket Layer                  │
│                    `main.go` (Gin router)                    │
├──────────────────┬──────────────────┬───────────────────────┤
│   Page Handlers  │  REST API Handlers│   WebSocket Handler   │
│  `controllers/  │  `controllers/    │  `controllers/        │
│   pages.go`     │   podcast.go`     │   websockets.go`      │
└────────┬─────────┴────────┬─────────┴──────────┬────────────┘
         │                  │                     │
         ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│              `service/podcastService.go`                     │
│              `service/fileService.go`                        │
│              `service/itunesService.go`                      │
│              `service/gpodderService.go`                     │
└──────────────────────────┬──────────────────────────────────┘
                           │
         ▼                 ▼                  ▼
┌─────────────────┐ ┌──────────────┐ ┌────────────────────────┐
│   Data Access   │ │  File System │ │  External APIs          │
│   `db/`         │ │  Downloads   │ │  iTunes, PodcastIndex,  │
│                 │ │  Backups     │ │  gPodder                │
└─────────────────┘ └──────────────┘ └────────────────────────┘
         │
         ▼
┌─────────────────┐
│  SQLite (GORM)  │
│  `podgrab.db`   │
└─────────────────┘
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

**Overall:** Monolithic MVC with server-side templates (Go html/template)

**Key Characteristics:**
- No separation between "handler" and "controller" — controller functions directly call service functions
- Service layer is flat (no interfaces except `SearchService` for search providers)
- Global mutable state: `db.DB` is a package-level `*gorm.DB` singleton initialized in `main.go`
- WebSocket state held in package-level maps (`activePlayers`, `allConnections`, `broadcast`)
- Cron-based background tasks scheduled in `main.go` via `gocron`
- HTML templates in `client/` mixed with inline JS/Vue.js
- Static assets in `webassets/` (vendored JS/CSS libraries, no build system)

## Layers

**Presentation Layer (controllers + client):**
- Purpose: Handle HTTP requests, render templates, manage WebSocket connections
- Location: `controllers/`, `client/`
- Contains: Gin handler functions, Go HTML templates, inline JavaScript
- Depends on: `service/`, `db/`, `model/`
- Used by: Gin router in `main.go`

**Service Layer:**
- Purpose: Business logic — podcast management, file I/O, search, OPML import/export
- Location: `service/`
- Contains: Pure functions and service structs (only `SearchService` interface + `ItunesService`/`PodcastIndexService` implementations)
- Depends on: `db/`, `model/`, `internal/sanitize`
- Used by: Controllers, cron jobs

**Data Layer (db):**
- Purpose: Data persistence, database queries, migrations
- Location: `db/`
- Contains: GORM model definitions, query functions, migration runner, job locking
- Depends on: GORM, SQLite driver
- Used by: Service layer, some controllers directly

**Model Layer:**
- Purpose: DTOs for external data formats (RSS, OPML, iTunes) and query parameters
- Location: `model/`
- Contains: XML/JSON deserialization structs, query/filter models, custom error types
- Depends on: Nothing (pure data structures)
- Used by: All other layers

## Data Flow

### Primary Request Path (Add Podcast)

1. POST `/podcasts` → `controllers.AddPodcast` (`controllers/podcast.go:402`)
2. `service.AddPodcast(url)` parses RSS feed → `service.FetchURL` (`service/podcastService.go:209`)
3. `db.CreatePodcast` + `db.CreatePodcastItem` persist to SQLite (`db/dbfunctions.go:297-305`)
4. Async `go service.RefreshEpisodes()` downloads episodes (`service/podcastService.go:617`)

### Episode Download Flow

1. Cron triggers `service.DownloadMissingEpisodes` (`service/podcastService.go:515`)
2. Job lock acquired via `db.GetLock` / `db.Lock` (`db/dbfunctions.go:329-352`)
3. Concurrent downloads with `sync.WaitGroup`, throttled by `MaxDownloadConcurrency` (`service/podcastService.go:531-546`)
4. Files saved to `$DATA/<sanitized-podcast-title>/` via `service.Download` (`service/fileService.go:25`)
5. Status updated via `service.SetPodcastItemAsDownloaded` → `db.UpdatePodcastItem` (`db/dbfunctions.go:311`)

### WebSocket Player Sync

1. Client connects to `/ws` → `controllers.Wshandler` upgrades connection (`controllers/websockets.go:34`)
2. Messages broadcast via `broadcast` channel to `HandleWebsocketMessages` goroutine (`controllers/websockets.go:65`)
3. Player state (register/unregister/enqueue) shared across connected clients

**State Management:**
- Database: SQLite via GORM (global `db.DB` variable)
- In-memory: WebSocket connection maps (`activePlayers`, `allConnections`) — not thread-safe for concurrent map writes
- Settings: `db.GetOrCreateSetting()` called per request in middleware — singleton row in SQLite

## Key Abstractions

**SearchService Interface:**
- Purpose: Abstract podcast search across providers
- Examples: `service/itunesService.go` (`ItunesService`, `PodcastIndexService`)
- Pattern: Strategy pattern via interface `SearchService` with `Query(q string) []*model.CommonSearchResultModel`

**GORM Models with UUID Base:**
- Purpose: All entities inherit UUID primary key + timestamps via `db.Base`
- Examples: `db.Podcast`, `db.PodcastItem`, `db.Setting`, `db.Tag`, `db.JobLock`
- Pattern: `db.Base` embeds `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`; `BeforeCreate` hook generates UUID

## Entry Points

**HTTP Server (`main.go`):**
- Location: `main.go:21`
- Triggers: Process start
- Responsibilities: DB init, Gin router setup, template loading, route registration, middleware, cron scheduling

**Cron Jobs (`main.go:intiCron`):**
- Location: `main.go:220`
- Triggers: On process start (background goroutine)
- Responsibilities: Periodic episode refresh, missing file check, image downloads, backup creation

**WebSocket Handler:**
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

**What happens:** Controllers in `controllers/podcast.go` call `db.*` functions directly (e.g., `db.GetPodcastById`, `db.GetAllPodcastItemsByPodcastId`, `db.GetAllTags`, `db.AddTagToPodcast`, `db.RemoveTagFromPodcast`)
**Why it's wrong:** Bypasses the service layer, creating inconsistent data access patterns. Some operations go through service (adding podcasts), others hit DB directly (retrieving podcasts by ID).
**Do this instead:** Route all data access through service-layer functions in `service/` to maintain consistency.

### Global Mutable State Without Thread Safety

**What happens:** `activePlayers` and `allConnections` maps in `controllers/websockets.go:22-23` are read/written from multiple goroutines without synchronization
**Why it's wrong:** Concurrent goroutine access to Go maps without mutex protection causes data races and potential panics
**Do this instead:** Use `sync.RWMutex` or `sync.Map` for shared state, or channel-based coordination

### Hardcoded API Keys

**What happens:** PodcastIndex API key and secret are hardcoded in `service/itunesService.go:42-44`
**Why it's wrong:** Credentials in source code are a security risk and cannot be rotated without code changes
**Do this instead:** Move to environment variables (already using `os.Getenv` pattern elsewhere) or configuration file

### Duplicate Controller Function

**What happens:** `DeletePodcasDeleteOnlyPodcasttEpisodesById` in `controllers/podcast.go:162` appears to be a typo-duplicated version of `DeletePodcastEpisodesById` — same logic, misspelled name
**Why it's wrong:** Dead code that confuses maintainers and is never routed to
**Do this instead:** Remove the duplicate function

## Error Handling

**Strategy:** Inconsistent — mix of ignored errors, `fmt.Println` logging, and `zap.Logger`

**Patterns:**
- Service functions return `error` but callers frequently ignore them (e.g., `controllers/podcast.go:196`: `go service.RefreshEpisodes()` ignores return)
- `fmt.Println` used for error logging in many places (e.g., `controllers/podcast.go:92-93`, `service/podcastService.go:276`)
- `zap` logger initialized in `service/podcastService.go:24-30` but rarely used
- Panic-on-error in `service/fileService.go:405-408` (`checkError` function panics)
- No structured error types beyond `model.PodcastAlreadyExistsError` and `model.TagAlreadyExistsError`

## Cross-Cutting Concerns

**Logging:** Dual system — `zap.SugaredLogger` (rarely used) and `fmt.Println`/`log.Println` (pervasively used). No consistent log levels.

**Validation:** Minimal. Request binding uses Gin's `ShouldBind*` but most handlers don't validate inputs beyond type coercion. No input sanitization on podcast URLs.

**Authentication:** Optional HTTP Basic Auth via `PASSWORD` environment variable in `main.go:135-143`. Single user, no role-based access.

---

*Architecture analysis: 2026-05-11*
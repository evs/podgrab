# Codebase Structure

**Analysis Date:** 2026-05-11

## Directory Layout

```
podgrab/
├── client/              # Go HTML templates (server-side rendered views)
├── controllers/         # HTTP handlers (Gin controllers)
├── db/                  # GORM models, DB access functions, migrations
├── docs/                # Documentation
├── images/              # Screenshot images for README
├── internal/            # Internal packages (not importable externally)
│   └── sanitize/        # File name sanitization library
├── model/               # Data transfer objects and request/response models
├── service/             # Business logic layer
├── webassets/           # Static frontend assets (vendored, no build step)
│   ├── fa/              # Font Awesome icons
│   ├── modal/           # Modal CSS
│   └── webfonts/        # Web fonts
├── .github/             # GitHub config
│   ├── ISSUE_TEMPLATE/  # Issue templates
│   └── workflows/       # CI workflows
├── main.go              # Application entry point and Gin router setup
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose configuration
├── go.mod               # Go module definition
└── go.sum               # Go dependency checksums
```

## Directory Purposes

**`client/`:**
- Purpose: Go HTML template files rendered server-side by `html/template`
- Contains: `.html` templates with embedded Vue.js for client-side interactivity
- Key files: `index.html`, `player.html`, `episodes_new.html`, `episodes.html`, `podcast.html`, `addPodcast.html`, `settings.html`, `navbar.html`, `commoncss.html`, `scripts.html`

**`controllers/`:**
- Purpose: HTTP request handlers (Gin controller functions)
- Contains: Route handler functions for both page renders and JSON/XML API endpoints
- Key files: `podcast.go` (REST API + RSS endpoints), `pages.go` (HTML page renders), `websockets.go` (WebSocket player sync)

**`db/`:**
- Purpose: Data persistence — GORM model definitions, CRUD query functions, migrations, job locking
- Contains: Struct definitions for all DB entities, database initialization, migration execution, query helper functions
- Key files: `podcast.go` (model structs), `dbfunctions.go` (query functions), `db.go` (init), `migrations.go`, `base.go` (UUID base)

**`model/`:**
- Purpose: Data transfer objects for external formats and query parameters
- Contains: RSS XML structs, iTunes API response structs, OPML structs, query filter models, custom error types
- Key files: `podcastModels.go`, `rssModels.go`, `itunesModel.go`, `opmlModels.go`, `queryModels.go`, `gpodderModels.go`, `errors.go`

**`service/`:**
- Purpose: Business logic and external service integration
- Contains: Podcast management, file download, search providers, OPML I/O, natural time formatting
- Key files: `podcastService.go` (core logic: add, refresh, download, delete), `fileService.go` (downloads, backups, images), `itunesService.go` (search providers), `gpodderService.go` (gPodder API), `naturaltime.go`

**`internal/sanitize/`:**
- Purpose: Vendored file name sanitization (not accessible to external packages)
- Key files: `sanitize.go`

**`webassets/`:**
- Purpose: Static frontend assets served directly (no build/compile step)
- Contains: Vendored JavaScript libraries (Vue.js, Axios, Luxon, Amplitude), CSS (Skeleton, Font Awesome), SVG icons
- Key libraries: `vue.js`, `axios.min.js`, `amplitude.min.js`, `luxon.min.js`, `vue-multiselect.min.js`, `vue-toasted.min.js`

**`images/`:**
- Purpose: Screenshot images for documentation (README)
- Contains: JPG screenshots demonstrating the UI

## Key File Locations

**Entry Points:**
- `main.go`: Application entry point — DB init, router setup, middleware, route registration, cron scheduling, template loading

**Configuration:**
- `main.go` (lines 135-143, 220-234): Env var config (`PASSWORD`, `CONFIG`, `DATA`, `CHECK_FREQUENCY`)
- `Dockerfile`: Container build config
- `docker-compose.yml`: Docker Compose deployment config
- `.env`: Environment variables (not committed — listed in `.gitignore`)

**Core Logic:**
- `service/podcastService.go`: Podcast add/refresh/download/delete workflow
- `service/fileService.go`: File download, backup, image processing
- `db/dbfunctions.go`: All database query functions
- `db/podcast.go`: GORM model definitions (Podcast, PodcastItem, Setting, Tag, JobLock)

**Testing:**
- No test files exist in the codebase

**Templates:**
- `client/index.html`: Home page
- `client/player.html`: Audio player
- `client/episodes_new.html`: Episode listing (Vue.js powered)
- `client/addPodcast.html`: Podcast search & add
- `client/settings.html`: Settings page
- `client/podcast.html`: Single podcast view
- `client/tags.html`: Tags management

## Naming Conventions

**Files:**
- Go source: `camelCase.go` — e.g., `podcastService.go`, `gpodderService.go`, `dbfunctions.go`
- Templates: `camelCase.html` — e.g., `addPodcast.html`, `episodes_new.html`
- Models split by domain: `podcastModels.go`, `rssModels.go`, `itunesModel.go`, `opmlModels.go`, `queryModels.go`, `gpodderModels.go`

**Packages:**
- `controllers`: HTTP handlers (flat, no sub-packages)
- `db`: Database models and access (flat, no sub-packages)
- `model`: DTOs (flat, organized by external service/format)
- `service`: Business logic (flat, no sub-packages)
- `internal/sanitize`: Vendored internal library

**Functions:**
- Go standard: `PascalCase` for exported, `camelCase` for unexported
- Database queries: `Get*`, `Create*`, `Update*`, `Delete*` prefix pattern in `db/`
- HTTP handlers: `Get*`, `Add*`, `Delete*`, `Patch*` prefix pattern in `controllers/`
- Service functions: Mixed — some are methods on structs (`ItunesService.Query`), most are package-level functions

**Variables:**
- Package-level globals: `PascalCase` — e.g., `db.DB`, `controllers.activePlayers`
- Local variables: `camelCase`

**Types:**
- Models: `PascalCase` — e.g., `PodcastAlreadyExistsError`, `EpisodesFilter`, `PodcastData`
- DTO structs: `PascalCase` — e.g., `SearchQuery`, `AddPodcastData`, `SettingModel`

## Where to Add New Code

**New Feature (API endpoint):**
- Handler: `controllers/podcast.go` — add handler function
- Route: `main.go` — register route in `router` group
- Service function: `service/podcastService.go` — add business logic
- DB function (if needed): `db/dbfunctions.go` — add query function

**New Feature (UI page):**
- Handler: `controllers/pages.go` — add page handler function
- Route: `main.go` — register GET route
- Template: `client/<name>.html` — create Go template file
- Add template reference in `main.go:130` template glob

**New Database Model:**
- Model struct: `db/podcast.go` — add GORM struct with `Base` embed
- Migration: `db/migrations.go` — add migration entry to `migrations` slice, add struct to `DB.AutoMigrate` call in `db/db.go:38`
- CRUD functions: `db/dbfunctions.go` — add Get/Create/Update/Delete functions

**New External Search Provider:**
- Interface implementation: `service/itunesService.go` — implement `SearchService` interface
- Registration: `controllers/pages.go` — add to `searchProvider` map

**New Background Task:**
- Service function: `service/` — add task function
- Cron registration: `main.go:intiCron` — add `gocron.Every(N).Do()` call

**New Query/Filter Parameter:**
- Model: `model/queryModels.go` — add fields to `EpisodesFilter` or create new filter struct
- DB query: `db/dbfunctions.go:57` — add filter conditions to `GetPaginatedPodcastItemsNew`
- Handler: `controllers/podcast.go` or `controllers/pages.go` — bind query params

## Special Directories

**`webassets/`:**
- Purpose: Static frontend assets (JS libraries, CSS, fonts, images)
- Generated: No — manually vendored
- Committed: Yes
- Note: No build system — all assets are pre-built vendor files

**`internal/`:**
- Purpose: Go internal packages (cannot be imported by external projects)
- Generated: No
- Committed: Yes

**`client/`:**
- Purpose: Go HTML template files parsed by `html/template`
- Generated: No
- Committed: Yes
- Note: Templates include substantial inline JavaScript (Vue.js apps); referenced by `main.go:130` glob pattern

**`images/`:**
- Purpose: README screenshots only
- Generated: No
- Committed: Yes

**`.github/`:**
- Purpose: GitHub-specific configuration (issue templates, CI workflows)
- Generated: No
- Committed: Yes

---

*Structure analysis: 2026-05-11*
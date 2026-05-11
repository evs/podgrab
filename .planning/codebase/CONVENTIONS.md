# Coding Conventions

**Analysis Date:** 2026-05-11

## Naming Patterns

**Packages:**
- Go standard: lowercase single-word package names
- Packages: `db`, `model`, `service`, `controllers`, `sanitize` (under `internal/`)
- Module: `github.com/akhilrex/podgrab`

**Files:**
- camelCase for multi-word files: `podcastService.go`, `fileService.go`, `itunesService.go`, `gpodderService.go`, `naturaltime.go`, `podcastModels.go`, `rssModels.go`, `queryModels.go`, `gpodderModels.go`, `podcastModels.go`, `dbfunctions.go`
- Single-word files for core concepts: `db.go`, `base.go`, `errors.go`

**Functions:**
- PascalCase for exported functions: `GetAllPodcasts`, `AddPodcast`, `FetchURL`, `DeleteFile`
- camelCase for unexported functions: `makeQuery`, `getSortOrder`, `createFolder`, `downloadImageLocally`, `deleteOldBackup`, `changeOwnership`, `getFileName`, `cleanFileName`, `checkError`
- Constructor-style not used — no `New*` factory functions (except `podcastindex.NewClient` from external lib)
- Verbose descriptive names: `SetPodcastItemAsQueuedForDownload`, `SetPodcastItemAsDownloaded`, `GetAllPodcastItemsByPodcastId`, `GetPaginatedPodcastItemsNew`

**Variables:**
- camelCase: `podcastItem`, `searchByIdQuery`, `latestDate`, `podcastListQuery`
- Short names in loops: `obj`, `v`, `item`, `pod`
- Map entries: `countMap`, `sizeMap`, `keyMap`
- Acronyms preserved: `URL`, `GUID`, `ID` — but inconsistently: `url`, `Id` (not `ID`)

**Types:**
- PascalCase structs: `Podcast`, `PodcastItem`, `Setting`, `Tag`, `JobLock`
- Custom error types with struct: `PodcastAlreadyExistsError`, `TagAlreadyExistsError`
- Enums via `iota`: `DownloadStatus` (`NotDownloaded`, `Downloading`, `Downloaded`, `Deleted`)
- String enums via typed constants: `EpisodeSort` (`RELEASE_ASC`, `RELEASE_DESC`, `DURATION_ASC`, `DURATION_DESC`)

## Code Style

**Formatting:**
- No linter or formatter config files detected (`.eslintrc`, `.prettierrc`, etc.)
- Go standard `gofmt` formatting assumed
- Mixed comment styles: some `//` with text, many `//` blank comments left as placeholders

**Linting:**
- No linting configuration detected
- No CI linting step in `.github/workflows/hub.yml` — only Docker build/push

**Key Code Smells:**
- `fmt.Println` used extensively for logging (62+ occurrences) instead of structured logger
- Inconsistent error handling: some `fmt.Println(err)`, some `log.Println(err)`, some `Logger.Errorw(...)`
- One `panic(err)` in `service/fileService.go:407` (`checkError` function)
- One `log.Fatal` in `service/itunesService.go:52` — crashes the process on PodcastIndex search error
- Commented-out `fmt.Println` statements left in production code throughout
- Typo in function name: `DeletePodcasDeleteOnlyPodcasttEpisodesById` (line 162, `controllers/podcast.go`)
- Typo in function name: `NatualTime` (should be `NaturalTime`, `service/naturaltime.go:9`)
- Global mutable state: `db.DB` (package-level `*gorm.DB`), `var Logger`, `var activePlayers`, `var allConnections`, `var broadcast`

## Import Organization

**Order:**
1. Standard library (`fmt`, `os`, `path`, `time`, etc.)
2. Third-party packages (`github.com/gin-gonic/gin`, `gorm.io/gorm`, etc.)
3. Internal packages (`github.com/akhilrex/podgrab/db`, `github.com/akhilrex/podgrab/service`, etc.)

**Path Aliases:**
- `stringy "github.com/gobeam/stringy"` — aliased import
- `strip "github.com/grokify/html-strip-tags-go"` — aliased import
- `uuid "github.com/satori/go.uuid"` — aliased import
- No explicit import grouping beyond what `goimports` would produce

## Error Handling

**Patterns:**
- `error` return values propagated from DB layer: `return result.Error` (`db/dbfunctions.go`)
- Custom error types implement `Error()` stringer: `model/errors.go`
- Type assertion for custom errors in controllers: `if v, ok := err.(*model.PodcastAlreadyExistsError); ok`
- `errors.Is(err, gorm.ErrRecordNotFound)` for GORM-specific checks
- Many errors silently consumed: `fmt.Println(err)` without returning or handling
- `ShouldBind*` / `ShouldBindQuery` / `ShouldBindUri` used for request validation — errors often result in `gin.H{"error": "Invalid request"}`

**Anti-patterns:**
- Errors printed but not returned: `fmt.Println(err)` after `db.GetPodcastById` (lines 92, 182, 195, 231 in `controllers/podcast.go`)
- Error swallowed in `makeQuery`: `body, _ := makeQuery(url)` in `service/gpodderService.go:19` and `service/itunesService.go:25`
- Mixed `fmt.Println` vs `log.Println` vs `Logger.Errorw` for same conceptual purpose

## Logging

**Framework:** `go.uber.org/zap` (sugared logger)

**Patterns:**
- Structured logging via `Logger.Errorw("message", err)` in select places (`service/podcastService.go`)
- Most of the codebase uses `fmt.Println` / `fmt.Printf` for debug output
- `log.Println` / `log.Print` used sparsely in `main.go` and `db/db.go`
- No log levels configured — `zap.NewProduction()` in init()

**When to Log:**
- Errors from HTTP requests: mixed `fmt.Println`, `log.Println`
- Debug info: `fmt.Println` scattered throughout cron jobs and websocket handlers
- Structured errors: `Logger.Errorw` only in `service/podcastService.go`

## Comments

**When to Comment:**
- Structural comments: `//PodcastData is`, `//PodcastItem is`, `//Base is`, `//BeforeCreate` — stub comments that add no value
- Many `//` blank line comments left behind (e.g., `//fmt.Println(...)` — commented-out debug)
- No doc comments on exported functions — Go convention would expect them
- Inline explanation for date parsing chain in `service/podcastService.go` (RFC1123 format variations)

**JSDoc/TSDoc:**
- Not applicable — Go project
- Go doc comments are mostly absent on exported symbols

## Function Design

**Size:** No enforced size limits. Some functions are very long:
- `AddPodcastItems` in `service/podcastService.go` ~110 lines
- `PlayerPage` in `controllers/pages.go` ~50 lines
- Template function map in `main.go` spans ~100 lines inline

**Parameters:** Multiple-parameter functions common:
- `UpdateSettings` takes 12 parameters (`service/podcastService.go:766`)
- `Download` takes 4 parameters (`service/fileService.go:25`)
- Pointer receivers used for mutation, value types for queries

**Return Values:**
- DB functions: `(result, error)` tuple convention
- Service functions: sometimes return pointer, sometimes value — inconsistent
- `GetOrCreateSetting` returns `*Setting` with no error (panics on failure silently handled via GORM)

## Module Design

**Exports:**
- All public symbols in a package are accessible (Go visibility rules)
- No `internal/` boundary enforcement except `internal/sanitize`
- `db` package leaks `DB` as exported global variable — accessible from all packages

**Barrel Files:**
- No explicit barrel/index files
- Each package exposes its types through individual files
- `db/base.go` defines shared `Base` struct used across all DB models

---

*Convention analysis: 2026-05-11*
# Codebase Concerns

**Analysis Date:** 2026-05-11

## Tech Debt

**Hardcoded API Credentials:**
- Issue: PodcastIndex API key and secret are hardcoded directly in source code
- Files: `service/itunesService.go` (lines 42-44)
- Impact: Anyone with access to the source or compiled binary can extract these credentials; rotating them requires code changes and recompilation
- Fix approach: Move `PODCASTINDEX_KEY` and `PODCASTINDEX_SECRET` to environment variables, loaded via `os.Getenv()` like other config values (PASSWORD, DATA, CONFIG)

**Global Mutable DB Singleton:**
- Issue: `db.DB` is a package-level global `*gorm.DB` pointer initialized in `db.Init()` and mutated in place; all packages access it directly
- Files: `db/db.go` (line 15), every file in `db/` and `service/` that references `db.DB`
- Impact: No dependency injection; impossible to use test doubles; concurrent access relies entirely on SQLite's locking; global state makes the app hard to reason about
- Fix approach: Encapsulate `*gorm.DB` in a repository struct, inject it via constructor functions, and pass it through the call stack

**Outdated Go Version and Dependencies:**
- Issue: `go.mod` specifies `go 1.15` (released August 2020), and many dependencies are severely outdated (e.g., `gorm.io/gorm v1.20.2` current is v1.25+, `github.com/dgrijalva/jwt-go v3.2.0` has known vulnerabilities)
- Files: `go.mod`
- Impact: Missing security patches, no access to modern Go toolchain features, known CVEs in JWT library
- Fix approach: Systematically upgrade dependencies with `go get -u`, starting with security-critical ones (jwt-go → `github.com/golang-jwt/jwt/v5`), then gorm, gin, and others

**Deprecated `ioutil` Usage:**
- Issue: Multiple files use `io/ioutil` which is deprecated since Go 1.16
- Files: `service/podcastService.go` (line 8), `service/fileService.go` (line 10), `service/fileService.go` (line 103)
- Impact: Compiler warnings, future removal will break builds
- Fix approach: Replace `ioutil.ReadAll` → `io.ReadAll`, `ioutil.WriteFile` → `os.WriteFile`

**No Structured Error Handling:**
- Issue: Errors are frequently logged with `fmt.Println` or silently ignored; error values from `db` functions are often discarded
- Files: `controllers/podcast.go` (lines 92, 183, 196), `service/podcastService.go` (lines 26-29), `service/gpodderService.go` (lines 25, 35, 44, 52)
- Impact: Silent failures make debugging extremely difficult; errors are swallowed leading to inconsistent state
- Fix approach: Establish consistent error handling: log errors with the structured logger, propagate errors up to the HTTP handler layer, and return appropriate HTTP status codes

## Known Bugs

**Typo in Controller Function Name:**
- Symptoms: Duplicate handler `DeletePodcasDeleteOnlyPodcasttEpisodesById` registered for `DELETE /podcasts/:id/podcast` — this appears to be a copy-paste error with a mangled name that calls `DeletePodcastEpisodes` instead of `DeleteOnlyPodcastById`
- Files: `controllers/podcast.go` (line 162)
- Trigger: Any `DELETE /podcasts/:id/podcast` request will delete episodes instead of just the podcast record
- Workaround: None — the route and handler are both wrong

**`removeStartingSlash` Template Function Adds Slash Instead of Removing:**
- Symptoms: The template function `removeStartingSlash` adds a slash when the input already starts with `/`, and adds one when it doesn't — the opposite of its stated purpose
- Files: `main.go` (lines 44-49)
- Trigger: Any template that calls `removeStartingSlash` produces incorrect URL paths
- Workaround: None

**`fmt.Println` Debug Output in Template Function:**
- Symptoms: The `removeStartingSlash` function calls `fmt.Println(raw)`, causing every template render to print the raw value to stdout
- Files: `main.go` (line 45)
- Trigger: Every page render that uses this function
- Workaround: None

**Database Init Error Swallowed:**
- Symptoms: If `db.Init()` fails, the error is printed with typo `statuse` but the program continues executing, leading to nil pointer panics on any DB access
- Files: `main.go` (line 25)
- Trigger: Database connection failure (e.g., invalid CONFIG path, permissions issue)
- Workaround: None — should `log.Fatal` or `os.Exit` on init error

**`DownloadMissingEpisodes` Concurrency Batching Bug:**
- Symptoms: The `wg.Wait()` inside the loop only blocks when `index % MaxDownloadConcurrency == 0`, meaning all other goroutines fire unconstrained; the batching logic is incorrect — it waits for the batch at the boundary but all goroutines between boundaries run in parallel
- Files: `service/podcastService.go` (lines 531-543)
- Trigger: Downloading more episodes than `MaxDownloadConcurrency` at once
- Workaround: MaxDownloadConcurrency defaults to 5, which limits practical impact

**Date Parsing Ignores Many Formats:**
- Symptoms: PubDate parsing tries 5 hardcoded formats sequentially, silently producing a zero-time value for unrecognized formats — episodes with non-standard dates get `time.Time{}` as their pub date
- Files: `service/podcastService.go` (lines 278-301)
- Trigger: Podcast feeds with ISO 8601, RFC 3339, or other non-RFC1123 date formats
- Workaround: None — episodes appear with missing/incorrect dates

## Security Considerations

**Hardcoded Secrets in Source Code:**
- Risk: PodcastIndex API credentials are committed to the repository in plain text
- Files: `service/itunesService.go` (lines 42-44)
- Current mitigation: None
- Recommendations: Move to environment variables; rotate the currently exposed keys; add secrets to `.gitignore`

**Basic Auth with Single User:**
- Risk: Authentication is a single hardcoded username "podgrab" with a password from the `PASSWORD` env var; no rate limiting, no HTTPS enforcement, no session management
- Files: `main.go` (lines 135-143)
- Current mitigation: Optional — if PASSWORD is empty, the app is completely unauthenticated
- Recommendations: Add rate limiting, support multiple users, enforce HTTPS, add session management

**No Input Validation on User-Supplied URLs:**
- Risk: The `AddPodcast` function takes a user-supplied URL and makes HTTP GET requests to it without validation; SSRF attacks are possible
- Files: `service/podcastService.go` (lines 209-243), `service/fileService.go` (lines 25-69)
- Current mitigation: None
- Recommendations: Validate URL scheme (http/https only), block private/internal IP ranges, set request timeouts

**WebSocket Origin Check Disabled:**
- Risk: The WebSocket upgrader uses default settings with no `CheckOrigin` — any origin can connect
- Files: `controllers/websockets.go` (lines 17-20)
- Current mitigation: None
- Recommendations: Add `CheckOrigin` handler to validate the origin header matches the server's domain

**Path Traversal Risk in File Operations:**
- Risk: File paths for downloads are constructed from podcast/episode titles sanitized via `sanitize.Name()`, but the `createFolder` function creates directories with `0777` permissions; `changeOwnership` only runs if PUID/PGID env vars are set
- Files: `service/fileService.go` (lines 361-369, 193-202), `internal/sanitize/sanitize.go`
- Current mitigation: Basic filename sanitization via `sanitize.Name()`
- Recommendations: Validate paths don't escape the data directory; use more restrictive directory permissions (0755)

**SQL Injection via User Sort Input:**
- Risk: The `GetAllPodcasts` function passes user-supplied `sorting` string directly to `DB.Order()`; similar risk in `GetPaginatedPodcastItemsNew` where `queryModel.Q` is used in a LIKE clause (though parameterized)
- Files: `db/dbfunctions.go` (line 29), `controllers/podcast.go` (lines 66-78)
- Current mitigation: The `PodcastPage` and other pages hardcode sort values; `GetAllPodcasts` API accepts user input
- Recommendations: Whitelist allowed sort columns and directions instead of passing raw user input

**Static File Serving Exposes Data Directory:**
- Risk: `router.Static("/assets", dataPath)` serves the entire DATA directory, potentially exposing downloaded podcast files and metadata
- Files: `main.go` (line 149)
- Current mitigation: Depends on correct DATA env var configuration
- Recommendations: Consider authenticated file serving; restrict to specific subdirectories

**`changeOwnership` Silently Fails:**
- Risk: If PUID/PGID env vars are invalid, `os.Chown` is skipped silently — files may be owned by root in Docker
- Files: `service/fileService.go` (lines 193-202)
- Current mitigation: None — only prints path to stdout
- Recommendations: Log warnings when ownership change fails; validate PUID/PGID on startup

## Performance Bottlenecks

**N+1 Query Pattern in GetAllPodcasts:**
- Problem: `GetAllPodcasts` loads all podcasts then calls `GetPodcastEpisodeStats` for aggregate stats, but each podcast's episodes are loaded via `Preload("Tags")` without preloading `PodcastItems`, requiring separate queries for stat computation
- Files: `service/podcastService.go` (lines 79-108), `db/dbfunctions.go` (lines 25-31)
- Cause: GORM's Preload doesn't aggregate efficiently; stats are computed in Go rather than SQL
- Improvement path: Use a single SQL query with JOINs and GROUP BY to compute all podcast stats at once

**Full Table Scan in GetEpisodeNumber:**
- Problem: Uses a RANK() window function with CTE that scans all episodes for a podcast every time an episode number is needed
- Files: `db/dbfunctions.go` (lines 251-269)
- Cause: No precomputed episode sequence; computed on every file download prefix calculation
- Improvement path: Pre-compute and store episode numbers during `AddPodcastItems` or cache results

**Dual Query Pattern for Paginated Results:**
- Problem: `GetPaginatedPodcastItemsNew` and `GetPaginatedPodcastItems` both execute the full query to get total count, then execute it again with LIMIT/OFFSET — two full table scans per page load
- Files: `db/dbfunctions.go` (lines 57-99, 101-126)
- Cause: GORM doesn't have built-in count+data in single query
- Improvement path: Use window functions (`COUNT(*) OVER()`) or raw SQL to get count and data in a single query

**Uncapped goroutines in OPML Import:**
- Problem: `AddOpml` launches a goroutine per outline without any semaphore or worker pool — a large OPML file could spawn hundreds or thousands of concurrent HTTP requests
- Files: `service/podcastService.go` (lines 117-141)
- Cause: No concurrency limit on OPML import
- Improvement path: Use a bounded worker pool (e.g., `semaphore.Weighted` or buffered channel) to limit concurrent podcast additions

**Debug Mode Left on in Production Query:**
- Problem: `GetPaginatedPodcastItemsNew` uses `DB.Debug()` which logs every SQL statement, and `TogglePodcastPauseStatus` also uses `DB.Debug()`
- Files: `db/dbfunctions.go` (line 60, line 277)
- Cause: Debug calls left in from development
- Improvement path: Remove `Debug()` calls or make them conditional on a debug flag

**SQLite Concurrency Limitations:**
- Problem: SQLite uses file-level locking; with concurrent downloads (goroutines) and cron jobs all writing to the same DB, write contention is inevitable at scale
- Files: `db/db.go`, `service/podcastService.go` (DownloadMissingEpisodes)
- Cause: SQLite's WAL mode is not explicitly enabled, and connection pool is set to 10 idle conns
- Improvement path: Enable WAL mode (`PRAGMA journal_mode=WAL`); consider migration to PostgreSQL for multi-user scenarios; add retry logic for locked DB errors

**HTTP Client Without Timeouts:**
- Problem: `makeQuery` and `httpClient` create `http.Client` without timeouts; `GetFileSizeFromUrl` uses `http.Head` without timeout — a slow or unresponsive server will block goroutines indefinitely
- Files: `service/podcastService.go` (lines 710-729), `service/fileService.go` (lines 255-273, 336-345)
- Cause: No timeout set on HTTP client or request context
- Improvement path: Set reasonable timeouts (e.g., 30s for requests, 60s for downloads) and use `context.WithTimeout`

**File Download Without Progress or Resume:**
- Problem: Downloads write entire file to disk in one `io.Copy` call; no resume support if interrupted; no progress tracking; response body closed after file close (defer order issue)
- Files: `service/fileService.go` (lines 25-68)
- Cause: Simple streaming download without chunking or resume headers
- Improvement path: Add `Range` header support for resume, progress callbacks via WebSocket, and proper defer ordering

## Fragile Areas

**Template Function `removeStartingSlash`:**
- Files: `main.go` (lines 44-49)
- Why fragile: Function logic is inverted (adds slash instead of removing), has a `fmt.Println` debug statement, and panics on empty string (accessing `raw[0]`)
- Safe modification: Rewrite function to actually remove starting slashes, remove debug print, add empty-string guard
- Test coverage: None

**WebSocket Connection Management:**
- Files: `controllers/websockets.go`
- Why fragile: Uses package-level maps (`activePlayers`, `allConnections`) with no mutex protection; concurrent reads/writes from multiple goroutines create race conditions; no cleanup on abnormal disconnect
- Safe modification: Add `sync.RWMutex` guards; implement proper connection cleanup; add ping/pong health checks
- Test coverage: None

**Cron Job Locking Mechanism:**
- Files: `db/dbfunctions.go` (lines 339-383), `service/podcastService.go` (lines 515-547)
- Why fragile: Locking uses DB records with time-based expiry checked by `UnlockMissedJobs`; no unique constraint on job name; potential for duplicate lock creation; `DownloadMissingEpisodes` acquires lock but the rest of the scheduled jobs don't
- Safe modification: Add unique constraint on `JobLock.Name`; acquire locks for all cron jobs; use transactions for lock check-and-create
- Test coverage: None

**RSS Feed Parsing:**
- Files: `model/podcastModels.go`, `model/rssModels.go`, `service/podcastService.go` (lines 245-356)
- Why fragile: RSS format is notoriously inconsistent across feeds; the parser uses a single struct that may not handle all namespace variations; date parsing has 5 hardcoded formats; missing fields default to zero values silently
- Safe modification: Add more date format attempts, handle missing GUIDs gracefully (fall back to URL), use a more robust RSS parsing library
- Test coverage: None

**Download Status State Machine:**
- Files: `db/podcast.go` (lines 69-76), `service/podcastService.go`
- Why fragile: `DownloadStatus` is an int iota (0-3) with no state transition validation; any status can be set to any other status; the `Downloading` status (1) is defined but never actually set — downloads go from `NotDownloaded` directly to `Downloaded`
- Safe modification: Implement explicit state transition validation; actually use the `Downloading` status during active downloads
- Test coverage: None

## Scaling Limits

**SQLite Single-Writer Limitation:**
- Current capacity: Works for personal use with a few dozen podcasts
- Limit: Single writer at a time; concurrent writes from downloads + cron jobs + API requests will see `database is locked` errors
- Scaling path: Enable WAL mode for better read concurrency; add retry with backoff for locked errors; long-term migration to PostgreSQL

**In-Memory WebSocket State:**
- Current capacity: Single-server only; all state is in Go maps
- Limit: No persistence; server restart loses all WebSocket connections; no horizontal scaling possible
- Scaling path: If multi-server is needed, use Redis Pub/Sub for WebSocket message broadcasting

**No Pagination on Many Queries:**
- Current capacity: Works with small podcast collections
- Limit: `GetAllPodcasts` loads all podcasts with all tags into memory; `GetAllPodcastItems` loads all items; `GetAllPodcastItemsToBeDownloaded` loads all non-downloaded items
- Scaling path: Add proper pagination with cursor-based navigation for all list endpoints; add filtering at DB level

**Global gocron Scheduler:**
- Current capacity: Single-process scheduling only
- Limit: If multiple instances run, cron jobs execute on each instance causing duplicate downloads and DB contention
- Scaling path: Use distributed locking (DB-based or Redis-based) before executing cron jobs; or move to a proper job queue

## Dependencies at Risk

**`github.com/dgrijalva/jwt-go v3.2.0+incompatible`:**
- Risk: This package has known CVEs and is unmaintained; the `+incompatible` suffix indicates it predates Go modules
- Impact: Not actually used for JWT signing in the codebase — appears to be a leftover dependency
- Migration plan: Remove if unused (check imports); if JWT is needed, migrate to `github.com/golang-jwt/jwt/v5`

**`github.com/jasonlvhit/gocron v0.0.1`:**
- Risk: Very old v0.0.1 release; minimal scheduling capabilities; no distributed locking support
- Impact: Cron jobs are the backbone of episode refresh/download automation
- Migration plan: Consider `github.com/go-co-op/gocron/v2` or a more robust scheduler with persistence

**`github.com/satori/go.uuid v1.2.0`:**
- Risk: This is the original satori UUID library which has known issues (non-RFC compliant in some cases); the community has forked to `github.com/google/uuid`
- Impact: Used for all primary keys in the database
- Migration plan: Switch to `github.com/google/uuid`; since UUIDs are stored as strings, migration is transparent

**`github.com/TheHippo/podcastindex v1.0.0`:**
- Risk: Third-party client for PodcastIndex API with minimal maintenance indicators
- Impact: Podcast search functionality depends on it
- Migration plan: Monitor for updates; consider direct HTTP API calls if library becomes unmaintained

## Missing Critical Features

**No Test Coverage:**
- Problem: Zero test files exist in the entire codebase
- Files: All packages
- Blocks: Refactoring, regression testing, CI validation, confident deployments

**No Authentication for WebSocket Endpoint:**
- Problem: The `/ws` route is registered on the unauthenticated root router, not inside the auth-protected group
- Files: `main.go` (lines 198-200)
- Blocks: Secure multi-user deployment; prevents unauthorized control of playback

**No HTTPS/TLS Support:**
- Problem: Gin server runs on plain HTTP; the app relies on a reverse proxy for TLS
- Files: `main.go` (line 206)
- Blocks: Secure deployment without a reverse proxy (common in home-lab setups)

**No Backup Restoration:**
- Problem: App can create backups but has no UI or API to restore from them
- Files: `service/fileService.go` (lines 275-303)
- Blocks: Recovery from data loss

**No Request Timeout Configuration:**
- Problem: HTTP client used for feed fetching and downloads has no timeout
- Files: `service/podcastService.go` (line 719), `service/fileService.go` (lines 336-345)
- Blocks: Reliable operation when external servers are slow or unresponsive

## Test Coverage Gaps

**Entire Service Layer:**
- What's not tested: All podcast service operations (add, refresh, download, delete)
- Files: `service/podcastService.go`, `service/fileService.go`
- Risk: Any refactoring could break core functionality silently
- Priority: High — core business logic

**Database Operations:**
- What's not tested: All CRUD operations, migrations, locking mechanism
- Files: `db/dbfunctions.go`, `db/db.go`, `db/migrations.go`
- Risk: SQL errors, constraint violations, race conditions go undetected
- Priority: High — data integrity

**HTTP Handlers:**
- What's not tested: All API endpoints, auth middleware, pagination, filtering
- Files: `controllers/podcast.go`, `controllers/pages.go`
- Risk: Route regressions, incorrect status codes, security vulnerabilities
- Priority: Medium

**RSS/OPML Parsing:**
- What's not tested: Feed parsing robustness, malformed input handling, date format variations
- Files: `model/podcastModels.go`, `service/podcastService.go` (AddPodcastItems)
- Risk: Broken feeds from various podcast hosts
- Priority: Medium

**WebSocket Communication:**
- What's not tested: Message routing, player registration, enqueue flow, connection cleanup
- Files: `controllers/websockets.go`
- Risk: Race conditions, memory leaks, message loss
- Priority: Low (auxiliary feature)

---

*Concerns audit: 2026-05-11*
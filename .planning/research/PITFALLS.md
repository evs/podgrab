# Domain Pitfalls

**Domain:** Self-hosted Go podcast manager modernization (Go 1.15 → 1.22+, GORM v1 → v2, Vue 2 → React SPA, adding REST API)
**Researched:** 2026-05-12

## Critical Pitfalls

Mistakes that cause rewrites, data loss, or extended outages.

### Pitfall 1: GORM v1→v2 Migration Breaks Existing Queries Silently

**What goes wrong:** GORM v2 has numerous breaking changes that compile without errors but produce different runtime behavior. The Podgrab codebase has several specific exposure points:
- `BlockGlobalUpdate` is now default — existing `Delete`/`Update` calls without `WHERE` clauses silently do nothing instead of affecting all rows (e.g., `DeletePodcastEpisodesById`, bulk status updates)
- Model tags changed from `snake_case` to `camelCase` — `auto_increment`, `unique_index`, `polymorphic_value` are silently ignored
- `db.CreateTable()` and `db.DropTable()` no longer exist; must use `db.Migrator()` interface
- `TableName()` return value is now cached — dynamic table names break silently
- `Soft Delete` requires `gorm.DeletedAt` type instead of just a field named `DeletedAt`
- Hook signatures changed from `func(u *User) error` to `func(tx *gorm.DB) error` — old hooks silently don't fire
- `Count()` only accepts `*int64` — existing `int` or `int32` calls fail silently
- Method chain safety: reusing a `*gorm.DB` after `.Where()` in goroutines causes data races

**Why it happens:** GORM v2 markets itself as "mostly compatible" which creates false confidence. The migration guide is comprehensive but developers skip reading it when `go build` succeeds.

**Consequences:** Data corruption (global updates that silently don't execute), lost hooks (business logic skipped), query race conditions in concurrent download goroutines.

**Prevention:**
1. Read the [GORM v2 Release Note](https://gorm.io/docs/v2_release_note.html) in full before any migration
2. Audit every model struct for `snake_case` struct tags — fix before touching `go.mod`
3. Search for all `db.Delete()` / `db.Update()` calls without `.Where()` — add explicit conditions or `AllowGlobalUpdate`
4. Search for all `Count()` calls and verify argument is `*int64`
5. Search for all hook methods and verify signature is `func(tx *gorm.DB) error`
6. Run `db.Session(&gorm.Session{})` before reusing `*gorm.DB` in goroutines
7. Write integration tests for all DB operations BEFORE upgrading GORM

**Detection:** After upgrade: queries return zero rows when they shouldn't; deletes don't delete; updates don't update; hooks don't fire; race detector (`go test -race`) flags data races in `db.Where(...)` chains.

**Phase:** Dependency upgrade phase — must be the FIRST thing validated after `go get -u gorm.io/gorm`

---

### Pitfall 2: gorilla/websocket Is Archived — Uncoordinated Replacement Breaks Real-Time Player

**What goes wrong:** The `gorilla/websocket` package was archived in 2023. Podgrab uses `gorilla/websocket v1.4.2` for real-time player sync across tabs (the `activePlayers` and `allConnections` maps in `controllers/websockets.go`). Projects that continue using archived packages face:
- No security patches for CVEs discovered after archive date
- No compatibility fixes for new Go versions
- Gradual bit-rot as Go runtime changes break subtle behavior

Meanwhile, the current WebSocket code already has critical concurrency bugs (bare map access from multiple goroutines with no mutex), and replacing the library + fixing concurrency simultaneously is a high-risk change.

**Why it happens:** "It works fine, we'll replace it later" — but later never comes, and when a Go version bump or security scan forces the change, it's done hastily without understanding the existing protocol.

**Consequences:** Complete loss of real-time player sync (the core differentiating feature); security vulnerabilities; race conditions causing panics.

**Prevention:**
1. Replace `gorilla/websocket` with `nhooyr.io/websocket` or `github.com/coder/websocket` (actively maintained, similar API)
2. Fix the mutex problem at the SAME time — wrap `activePlayers` and `allConnections` with `sync.RWMutex`
3. Document the WebSocket message protocol (message types: `RegisterPlayer`, `PlayerRemoved`, `Enqueue`, `Register`, `PlayerExists`, `NoPlayer`) before touching anything
4. Keep the existing message format unchanged so the Vue 2 client continues working during migration
5. Add WebSocket integration tests before replacing the library

**Detection:** Race detector flags concurrent map access; WebSocket connections silently drop; player sync stops working intermittently.

**Phase:** Concurrency safety phase (before React migration starts, since React app will also need WebSocket support)

---

### Pitfall 3: Dual-Frontend State Divergence During Incremental Migration

**What goes wrong:** During the Vue 2 → React incremental migration, the app has TWO frontends consuming the SAME backend API. The Vue 2 frontend lives in Go templates (e.g., `client/index.html` embeds Vue with `{{ .podcasts }}` server-rendered data). The React frontend will consume JSON from a new REST API. If these two APIs return different data shapes or different behavior for the same operation:
- User does action in Vue UI → sees one result
- User does same action in React UI → sees different result
- Tags, episode counts, download status get out of sync between views

The current architecture makes this especially dangerous because:
- Some endpoints return HTML pages (e.g., `GET /podcasts/:id/view` renders a template)
- Some endpoints return JSON (e.g., `GET /podcasts` returns JSON already)
- The same Gin route handles both `GET /podcasts` (returns JSON for all podcasts) and `GET /` (renders HTML page)
- Controllers sometimes bypass the service layer and hit `db.DB` directly (e.g., inline GORM calls in handlers)

**Why it happens:** Adding a REST API alongside existing HTML endpoints feels like "just adding JSON responses" but the existing controller logic is deeply entangled with template rendering. Adding a new API endpoint that calls slightly different service code produces subtly different results.

**Consequences:** Users lose trust in the app; data appears inconsistent; bugs reported against one UI can't be reproduced in the other; debugging becomes a nightmare.

**Prevention:**
1. **Define the canonical API FIRST** — write an OpenAPI spec for all REST endpoints before writing any React code
2. **Make existing HTML endpoints call the same API handlers** — route refactoring: `GET /api/podcasts` (JSON) and `GET /podcasts` (HTML) both call `controllers.GetAllPodcasts` which returns data, and the HTML version wraps it in a template
3. **Separate concerns**: API handlers return structured data; page handlers render templates from API data
4. **Integration tests that assert both endpoints return the same business data**
5. **Phase the migration by ROUTE, not by feature** — migrate all endpoints for `/podcasts/*` before starting `/episodes/*`

**Detection:** Same action produces different results in Vue vs React; tags appear in one UI but not the other; episode counts differ.

**Phase:** REST API phase (before building React UI components)

---

### Pitfall 4: Upgrading Go 1.15→1.22+ Exposes Latent Bugs in `ioutil` and Build Tags

**What goes wrong:** Go 1.15 → 1.22 spans 7 major releases. Several specific breaking changes affect this codebase:
- `io/ioutil` is deprecated since Go 1.16 and removed in Go 1.22+ — `ioutil.ReadAll`, `ioutil.WriteFile`, `ioutil.ReadFile` calls in `service/podcastService.go`, `service/fileService.go` won't compile
- `go mod` behavior changed significantly — `go 1.15` in `go.mod` means the module is in pre-module compatibility mode
- `dgrijalva/jwt-go v3.2.0+incompatible` — the `+incompatible` flag means it predates Go modules; newer Go toolchains handle this differently
- `golang.org/x/crypto` and `golang.org/x/net` at those old versions have known CVEs that are enforced at build time by some Go toolchain features
- New `go vet` checks will flag issues that previously compiled fine (e.g., `fmt.Println` with wrong format strings)

**Why it happens:** Developers upgrade Go version expecting a clean build, then spend days fixing compilation errors. They may be tempted to "just fix the compilation errors" without understanding behavioral changes.

**Consequences:** Build failures; subtle behavioral differences in crypto/network packages; CVEs remain even after "upgrading" because only the Go version was bumped, not the transitive dependencies.

**Prevention:**
1. Upgrade Go version FIRST, before any other dependency changes — isolate the blast radius
2. Replace all `io/ioutil` imports: `ioutil.ReadAll` → `io.ReadAll`, `ioutil.WriteFile` → `os.WriteFile`, `ioutil.ReadFile` → `os.ReadFile`
3. Run `go mod tidy` after the Go version change to clean up the module graph
4. Run `go vet` on the ENTIRE codebase after the version bump and fix all warnings
5. Specifically audit `fmt.Println` format strings (the WebSocket code has `fmt.Println("Failed to set websocket upgrade: %+v", err)` which is wrong — `Println` doesn't support format verbs)
6. Replace `dgrijalva/jwt-go` with `github.com/golang-jwt/jwt/v5` using the [migration guide](https://github.com/golang-jwt/jwt/blob/main/MIGRATION_GUIDE.md) — or remove it if unused as the CONCERNS.md suggests
7. Run `go test -race` after the Go version bump

**Detection:** Compilation errors, `go vet` warnings, `fmt.Println` producing literal `%+v` in output instead of formatted values.

**Phase:** First phase — Go version upgrade must be isolated from all other changes

---

### Pitfall 5: Global `db.DB` Singleton Prevents Testing and Introduces Race Conditions

**What goes wrong:** The `db.DB` package-level global `*gorm.DB` is accessed from every package. During modernization, this creates two problems:
1. **Impossible to write tests** — can't inject a test database, can't mock DB responses, can't run tests in parallel
2. **SQLite race conditions get worse** — concurrent goroutines (cron jobs, downloads, API requests) all write through the same connection pool. The current code doesn't enable WAL mode, so SQLite's file-level locking causes "database is locked" errors under load

When you add the REST API (which adds more concurrent HTTP handlers hitting the DB), race conditions will increase. When you add React (which makes more API calls per page load than the server-rendered pages did), concurrency spikes further.

**Why it happens:** Refactoring away from a global singleton feels like "later" work, but every new feature (REST API endpoints, React components) adds more concurrent DB access through the same unsafe path. Each new endpoint makes the problem worse.

**Consequences:** "database is locked" errors in production; test suite can't be built; refactors are unsafe because any change might break unrelated code using `db.DB`.

**Prevention:**
1. **Create a `Repository` struct** that wraps `*gorm.DB` and inject it via constructor functions:
   ```go
   type Repository struct {
       db *gorm.DB
   }
   func NewRepository(db *gorm.DB) *Repository { ... }
   ```
2. **Migrate packages one at a time** — start with `db/dbfunctions.go` (make methods on Repository), then `service/`, then `controllers/`
3. **Enable SQLite WAL mode immediately** — add `PRAGMA journal_mode=WAL` in `db.Init()`:
   ```go
   db.Exec("PRAGMA journal_mode=WAL")
   db.Exec("PRAGMA busy_timeout=5000")
   ```
4. **Set `MaxOpenConns` to 1 for SQLite writes** — SQLite only supports one writer; use `db.DB().SetMaxOpenConns(1)` or use a separate write connection
5. **Write DB-layer integration tests FIRST** using the injected repository before refactoring service/controller layers

**Detection:** "database is locked" errors in logs during concurrent operations; test failures due to shared state; race detector flags concurrent access to `db.DB`.

**Phase:** Early phase (before adding REST API endpoints), since every new API route makes the problem worse

---

## Moderate Pitfalls

### Pitfall 6: `removeStartingSlash` Template Function Bug Leaks Into API URLs

**What goes wrong:** The `removeStartingSlash` function in `main.go` (lines 44-49) is inverted — it ADDS a slash instead of removing one. It also calls `fmt.Println(raw)` on every template render. When building the REST API, this template function is irrelevant (React doesn't use Go templates), BUT any server-side code that was built to work WITH this bug (expecting the inverted behavior) will break if you fix it without understanding all callers.

**Why it happens:** The bug has existed since initial development, so all Go template code was written to work with the inverted behavior. Fixing the function breaks templates that depend on the current behavior.

**Consequences:** Broken URL paths in server-rendered pages; API URLs with doubled or missing slashes; stdout pollution in production logs.

**Prevention:**
1. Search ALL template files for `removeStartingSlash` usage — document what each call expects
2. Fix the function AND simultaneously fix any templates that relied on the inverted behavior
3. Remove the `fmt.Println(raw)` debug line
4. Add a guard for empty string (current code panics with index out of range on `raw[0]`)

**Detection:** Template rendering panics on empty strings; URL paths have leading double slashes in server-rendered pages; stdout log pollution.

**Phase:** Bug fix phase (early, but after Go version upgrade)

---

### Pitfall 7: React Build System Collides With Go Binary Deployment

**What goes wrong:** Podgrab is deployed as a single Go binary in a Docker container with Alpine. There is no npm, no build system, no Node.js — Vue 2 components are loaded as vendored JS files from `/webassets/`. Adding a React SPA requires:
- A Node.js build step to produce the React bundle
- A way to serve the built React bundle from the Go binary or a separate static file server
- Docker image changes to include the Node.js build step or a multi-stage build

The trap is either:
- **Embedding the React build into the Go binary** using `go:embed` — works, but the Dockerfile must include a Node.js build stage, and every backend change requires a full frontend rebuild
- **Serving React from a separate container** — adds deployment complexity, CORS issues, proxy configuration
- **Skipping the build system** — copying the pattern of vendored JS (like current webassets) — impossible with React, which requires JSX compilation and bundling

**Why it happens:** The current "just a Go binary" deployment model is simple and pleasant. React fundamentally requires a build step, and teams underestimate the build/deploy pipeline integration.

**Consequences:** Broken Docker builds; oversized Docker images; CORS errors in production; dev/prod environment differences cause React to fail in Docker but work locally.

**Prevention:**
1. **Use multi-stage Docker build**: Stage 1 = Node.js + `npm run build`, Stage 2 = Go build, Stage 3 = Alpine runtime with Go binary + built React files
2. **Use `go:embed`** to embed the React build output into the Go binary — eliminates runtime file serving complexity
3. **Configure Vite** (not webpack) for the React build — faster builds, simpler config, first-class proxy support for dev mode
4. **In dev mode, run Vite dev server** with proxy to Go backend (`/api/*` → `localhost:8080`) — in prod, serve embedded static files
5. **Make the Go server serve `index.html` for unknown routes** (SPA fallback) so React Router works

**Detection:** Docker build fails at Node.js stage; 404 on React routes after page refresh; CORS errors when React dev server calls Go API.

**Phase:** Infrastructure setup phase (before first React component is built)

---

### Pitfall 8: WebSocket Protocol Must Support Both Vue 2 and React Clients Simultaneously

**What goes wrong:** The WebSocket server in `controllers/websockets.go` uses a custom message protocol (`RegisterPlayer`, `Enqueue`, `Register`, `PlayerExists`, `NoPlayer`, `PlayerRemoved`). During incremental migration:
- The Vue 2 player page (`/player`) connects to `ws://host/ws` and sends `RegisterPlayer`
- The React player page (later) will connect to the same endpoint
- If the protocol changes to support React features, the Vue 2 client breaks
- If the protocol stays the same, React features are limited

The current protocol has no versioning, no message schema validation, and no documentation. The Vue 2 client uses custom delimiters (`${`, `}`) in template strings and relies on `Amplitude.js` for audio playback — the React client will use a different audio library.

**Why it happens:** The WebSocket protocol was never designed — it just grew. Adding a second client forces protocol decisions that were never made explicitly.

**Consequences:** Player sync breaks between Vue and React tabs; messages are malformed; one client's disconnect crashes the other.

**Prevention:**
1. **Document the exact WebSocket message format** — message types, field names, expected values
2. **Add a `version` or `clientType` field** to the `RegisterPlayer` message — server can branch behavior
3. **Keep backward compatibility** — never remove existing message types, only add new ones
4. **Build the React client to speak the SAME protocol first** — get it working identically, then add enhanced features
5. **Add integration tests for WebSocket message round-tripping** using both client protocol versions

**Detection:** "Player not found" errors when Vue and React clients are open simultaneously; message deserialization errors in server logs.

**Phase:** React player migration phase (not before — document the protocol earlier, but actual dual-client support only matters when React player exists)

---

### Pitfall 9: Server-Rendered Data in Go Templates vs. API-Fetched Data in React Creates Inconsistency

**What goes wrong:** The current Vue 2 frontend receives initial data embedded in Go templates — e.g., `client/index.html` line 814: `allPodcasts: {{ .podcasts }}`. This means:
- Data is available instantly on page load (no API call)
- Data is shaped by Go template marshaling (which may differ from JSON API serialization)
- Data includes computed fields from Go template functions (like `downloadedEpisodes`, `latestEpisodeDate`)

When the React frontend fetches from `/api/podcasts`, it gets a different JSON shape because:
- The template-rendered data uses Go's default `json.Marshal` behavior
- Template functions like `formatDate`, `naturalDate`, `formatFileSize` are computed server-side in templates
- The REST API endpoints return raw model data without these computed fields
- Episode counts (`DownloadedEpisodesCount`, `AllEpisodesCount`) are loaded via separate queries and attached in Go code, not included in the raw model

**Why it happens:** Template rendering and API serialization go through different code paths. Template functions pre-process data for human display; API endpoints return machine-readable data. The "same" data looks different.

**Consequences:** React UI shows different episode counts, different date formats, missing aggregate data; users see different information in Vue vs React.

**Prevention:**
1. **Define a single `PodcastDTO` struct** that represents what the UI needs (including computed fields like `DownloadedEpisodesCount`, `LastEpisode`, `DownloadedEpisodesSize`)
2. **Both template rendering and API endpoints use the same DTO** — template functions compute values into the DTO, API serializes the DTO
3. **Write a single service-layer function** that assembles the DTO from DB data — both controllers call it
4. **Integration test**: assert `GET /api/podcasts` JSON matches the data embedded in `GET /` template

**Detection:** Episode counts differ between Vue and React; date formats are inconsistent; aggregate stats missing in React.

**Phase:** REST API design phase (before building React components)

---

### Pitfall 10: `GetAllPodcasts` SQL Injection via User Sort Input Persists Into API

**What goes wrong:** `db/dbfunctions.go` line 29 passes user-supplied `sorting` string directly to `DB.Order(sorting)`. Currently this is only called from template-rendered pages that hardcode sort values. When you expose `GET /api/podcasts?sort=created_at` to the React frontend, any user can inject arbitrary SQL through the sort parameter (e.g., `sort=id; DROP TABLE podcasts`).

**Why it happens:** The existing code "works" because the only callers are server-side templates with hardcoded sort values. Adding an API endpoint that accepts arbitrary user input makes the vulnerability exploitable.

**Consequences:** SQL injection; data exfiltration; data destruction.

**Prevention:**
1. **Whitelist allowed sort columns and directions** — define a map of safe sort values to actual column names:
   ```go
   var allowedSorts = map[string]string{
       "name": "title", "dateadded": "created_at", 
       "lastepisode": "last_episode", "episodes": "episode_count",
   }
   ```
2. **Validate and translate** any user-supplied sort value through the whitelist before passing to GORM
3. **Apply same validation** to all other endpoints that accept user input for ordering or filtering

**Detection:** Sending `sort=id;SELECT 1` to the API returns unexpected results or errors; security scanner flags raw string in `ORDER BY`.

**Phase:** REST API phase (MUST fix before ANY endpoint accepts user-supplied sort/filter params)

---

## Minor Pitfalls

### Pitfall 11: `fmt.Println` Debug Statements in Production

**What goes wrong:** The codebase has `fmt.Println` calls throughout (WebSocket handler, template functions, service layer). These:
- Write to stdout without timestamps or structure
- Cannot be filtered, searched, or disabled
- Pollute Docker container logs
- The `removeStartingSlash` function prints EVERY template render's input

**Why it happens:** No logging discipline was established. `fmt.Println` is the path of least resistance.

**Consequences:** Unsearchable logs in production; performance impact from excessive printing; critical errors invisible among debug noise.

**Prevention:**
1. Replace ALL `fmt.Println` with structured `zap.SugaredLogger` calls (the project already imports zap, it's just rarely used)
2. Define log levels: Debug for development, Info for operational events, Error for failures
3. Remove `fmt.Println(raw)` from `removeStartingSlash` — this is pure debug noise
4. Configure zap to log JSON in Docker for structured log collection

**Detection:** Docker container logs are noisy and unstructured; `grep` can't find errors; log volume is excessive.

**Phase:** Logging/cleanup phase (can be done incrementally alongside other work)

---

### Pitfall 12: `satori/go.uuid` vs `google/uuid` UUID Generation Differences

**What goes wrong:** The codebase uses `github.com/satori/go.uuid v1.2.0` for all primary keys. The satori library has known issues:
- Non-RFC-compliant UUID v4 generation in some cases
- The library is unmaintained and archived
- Migration to `github.com/google/uuid` is recommended but generates slightly different UUIDs

Since UUIDs are stored as strings in SQLite, the migration is transparent at the DB level — existing UUIDs remain valid. But during migration, if both libraries are imported simultaneously (during gradual replacement), they may generate different UUID formats that look inconsistent.

**Why it happens:** UUID generation seems trivial and library replacement seems safe, but subtle format differences can cause issues.

**Consequences:** Minimal — existing data is unaffected, new UUIDs may have subtly different format but are still valid UUIDs.

**Prevention:**
1. Replace `satori/go.uuid` with `google/uuid` — it's a near-drop-in replacement
2. Do it in ONE commit, not gradually — avoid having both libraries imported
3. Verify no code depends on satori-specific features (e.g., `uuid.NewV1` — google/uuid uses `NewUUID` with version parameter)

**Detection:** UUID format differences in new records; build errors if both packages imported.

**Phase:** Dependency upgrade phase (minor, do alongside other dep upgrades)

---

### Pitfall 13: OPML Import Uncapped Goroutines Will Get Worse With API Usage

**What goes wrong:** `AddOpml` launches a goroutine per outline without any concurrency limit. A large OPML file could spawn hundreds of concurrent HTTP requests. When the API exposes OPML import, automated tools or scripts could trigger this with even larger files.

**Why it happens:** The current cron-triggered import doesn't face OPML files with hundreds of entries in practice. Adding an API endpoint makes it accessible to automation.

**Consequences:** API timeouts, resource exhaustion, IP bans from podcast directories due to excessive requests.

**Prevention:**
1. Use `semaphore.Weighted` or a buffered channel as a worker pool
2. Limit concurrent podcast additions to a configurable number (e.g., 5)
3. Return a progress indicator via WebSocket or polling endpoint
4. Same fix applies to `DownloadMissingEpisodes` concurrency batching bug

**Detection:** Server becomes unresponsive during OPML import; external API rate limiting.

**Phase:** Concurrency safety phase

---

### Pitfall 14: React Router vs Gin Router Conflict

**What goes wrong:** React Router handles client-side routing (e.g., `/podcasts/123`, `/episodes`, `/settings`). Gin handles server-side routing. When a user refreshes a React page at `/podcasts/123`, Gin doesn't have a route for that path and returns 404. Currently Gin routes like `GET /podcasts/:id/view` render Go templates — these routes will need to either:
- Continue serving the template for backwards compatibility during migration
- Be replaced with a "serve React index.html" catch-all
- Or split between API routes (`/api/*`) and page routes

**Why it happens:** SPA routing and server routing are fundamentally different. The server needs to serve `index.html` for any route the SPA handles, but also needs to keep serving API routes and existing template routes.

**Consequences:** 404 on page refresh in React; broken deep links; Vue 2 pages accidentally catch React routes.

**Prevention:**
1. **Namespace all new API routes under `/api/`** — `/api/podcasts`, `/api/episodes`, `/api/tags`
2. **Add a catch-all route** at the END of Gin routing that serves the React `index.html` for any non-API, non-asset route
3. **Migrate one page at a time** — when `/episodes` React page is ready, remove the Gin route for `GET /episodes` and let the catch-all serve React
4. **Order matters**: API routes → asset routes → Vue 2 page routes (remaining) → catch-all for React

**Detection:** 404 on refreshing a React page; Vue 2 pages served when React page was expected.

**Phase:** React infrastructure setup phase

---

### Pitfall 15: Cron Job Duplicate Execution If Multiple Instances Run

**What goes wrong:** The `gocron` scheduler runs in-process. If a user starts two Podgrab containers (misconfiguration, rolling update, etc.), both instances run the same cron jobs, causing:
- Duplicate episode downloads
- Duplicate file size calculations
- DB contention from parallel writes

The existing `JobLock` mechanism only checks DB-level locks and has no unique constraint on job names, so duplicate lock creation is possible.

**Why it happens:** Single-instance assumption was never enforced. Home-lab users commonly experiment with multiple containers.

**Consequences:** Wasted bandwidth, duplicate file downloads, DB corruption from concurrent writes.

**Prevention:**
1. Add unique constraint on `JobLock.Name` column
2. Use `INSERT ... ON CONFLICT DO NOTHING` for lock acquisition
3. Consider replacing `github.com/jasonlvhit/gocron v0.0.1` with `github.com/go-co-op/gocron/v2` which has better distributed support
4. Add advisory locking at the application level (e.g., file lock or DB singleton check on startup)

**Detection:** Duplicate download directories; double the expected API calls to podcast directories; "database is locked" errors in logs.

**Phase:** Dependency upgrade phase (when replacing `gocron`)

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Go version upgrade (#4) | `ioutil` compilation errors, `fmt.Println` format bugs, `+incompatible` module handling | Upgrade Go FIRST in isolation; fix compilation before touching any deps |
| GORM v1→v2 migration (#1) | Silent behavioral changes in queries, hooks, tags, soft delete, BlockGlobalUpdate | Read release notes fully; integration tests before upgrade; audit all model tags |
| Dependency upgrades (jwt, uuid, websocket, gocron) | Multiple library changes at once make it impossible to identify which broke something | Upgrade ONE dependency at a time; test between each |
| Global `db.DB` refactoring (#5) | Tests can't be written without DI; SQLite "database is locked" escalates with new API | Inject repository before adding API routes; enable WAL mode; write tests first |
| REST API design (#3, #9, #10) | API returns different data shape than template rendering; SQL injection via sort params | Define DTOs; whitelist sort columns; integration tests for data parity |
| React build infrastructure (#7, #14) | No build system exists; Docker needs Node.js stage; SPA routing conflicts with Gin | Multi-stage Docker; `go:embed`; API namespacing; catch-all route |
| WebSocket protocol (#2, #8) | Archived gorilla/websocket; no mutex on maps; protocol not documented; dual-client support | Replace library + add mutex; document protocol; add version field; integration tests |
| Vue→React page migration (#3) | Both UIs show different data for same operations; state divergence | Migrate by route; both UIs call same service layer; test data parity |
| Cron job safety (#15) | Duplicate job execution; no unique constraint on JobLock.Name | Add unique constraint; replace gocron; add advisory locking |
| Production logging (#11) | Unstructured `fmt.Println` makes debugging impossible | Replace with zap; structured JSON logs; log levels |

## Sources

- [GORM v2 Release Note (Breaking Changes)](https://gorm.io/docs/v2_release_note.html) — HIGH confidence, official documentation
- [golang-jwt Migration Guide](https://github.com/golang-jwt/jwt/blob/main/MIGRATION_GUIDE.md) — HIGH confidence, official repository
- [Martin Fowler — Strangler Fig Application](https://martinfowler.com/bliki/StranglerFigApplication.html) — HIGH confidence, established pattern
- [gorilla/websocket documentation (Context7)](https://context7.com/gorilla/websocket/llms.txt) — HIGH confidence, concurrency requirements for WebSocket connections
- [Podgrab codebase analysis](.planning/codebase/CONCERNS.md) — HIGH confidence, direct codebase inspection
- [Go 1.16 Release Notes (ioutil deprecation)](https://go.dev/doc/go1.16) — HIGH confidence, official Go release notes
- React incremental migration patterns — MEDIUM confidence, React.dev docs focus on SSR hydration rather than server-template-to-SPA migration

---
*Pitfalls research: 2026-05-12*
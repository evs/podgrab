# Phase 2: Test Framework & Code Quality - Context

**Gathered:** 2026-05-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Set up a working Go test harness (`go test ./...` passes), fix three trivial code bugs, and add initial test coverage for core podcast CRUD operations. The codebase currently has **zero test files**. No new features are added — this phase only fixes what's broken and verifies existing behavior with tests.

**In scope:**
- Test harness: `go test ./...` must compile and run (empty test files are fine initially)
- Bug fixes: `DeletePodcasDeleteOnlyPodcasttEpisodesById` typo handler (controllers/podcast.go:162), `removeStartingSlash` reverse logic (main.go:44-50), debug `fmt.Println` in `removeStartingSlash` (main.go:45)
- Test coverage: DB layer and service layer tests for core podcast CRUD (add, get, delete)

**Out of scope:**
- Comprehensive test coverage across all layers (Phase 3+)
- WebSocket controller tests
- File service tests
- OPML/gPodder integration tests
- Error handling modernization (Phase 4)

</domain>

<decisions>
## Implementation Decisions

### Testing Framework
- **D-01:** Use Go's built-in `testing` package only — no testify, no external test dependencies. Rationale: Phase 1 already upgraded many dependencies; keep test tooling minimal and dependency-free to match the project's conservative dependency philosophy.
- **D-02:** `t.Fatal` / `t.Fatalf` for setup failures, plain `t.Errorf` for assertion failures. No helper library — just Go stdlib.

### Test Database
- **D-03:** Use SQLite `:memory:` database for all tests. GORM supports this natively with `gorm.Open(sqlite.Open(":memory:"), ...)`. Fast, isolated per-test, no file cleanup needed.
- **D-04:** Initialize DB and run migrations in `TestMain(m *testing.M)` (or a shared setup helper), so each test gets a fresh schema. Defer cleanup or use `t.Cleanup` to drop data between tests.

### Bug Fixes
- **D-05:** `DeletePodcasDeleteOnlyPodcasttEpisodesById` is a duplicate of `DeletePodcastEpisodesById` — **remove the typo function entirely**, keep the correctly-named one. Verified: both have identical body (lines 151-160 vs 162-171 in controllers/podcast.go).
- **D-06:** `removeStartingSlash` logic is reversed (adds "/" when no slash, returns unchanged when slash present). **Fix:** remove leading "/" when present, return unchanged otherwise. Also remove the debug `fmt.Println(raw)` on line 45 of main.go.

### Test Coverage Targets
- **D-07:** Minimum test files: `db/db_test.go` (DB CRUD functions), `service/podcastService_test.go` (service-layer functions). These prove the test harness works and cover the code paths touched by bug fixes.
- **D-08:** Tests must verify actual database behavior (read-after-write, delete-then-read-empty) — not just "function doesn't panic" smoke tests.

### the agent's Discretion
- Specific test function naming conventions (follow Go stdlib: `TestAddPodcast`, `TestGetAllPodcasts`)
- Whether to use table-driven tests or individual test functions (agent decides based on complexity)
- How to structure `db/db_test.go` test helpers (shared setup/teardown patterns)
- Exact assertions to use (reflect.DeepEqual vs manual field comparison — agent decides for readability)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase 1 Decisions (locked)
- `.planning/phases/01-dependency-upgrades/01-CONTEXT.md` — Prior dependency upgrade decisions, Go 1.24 + GORM v1.26 locked versions
- `.planning/STATE.md` — Accumulated project state and recent decisions

### Code to Fix
- `controllers/podcast.go` — Duplicate typo handler at line 162 (`DeletePodcasDeleteOnlyPodcasttEpisodesById`)
- `main.go` — `removeStartingSlash` template function at line 44-50 (reversed logic + debug print)

### Code to Test
- `db/dbfunctions.go` — DB CRUD: `GetPodcastByURL`, `GetAllPodcasts`, `GetPodcastById`, `AddPodcast`, `DeletePodcast`
- `service/podcastService.go` — Service layer: `AddPodcast`, `GetAllPodcasts`, `DeletePodcast`, `DeletePodcastEpisodes`
- `db/db.go` — `Init()` and `Migrate()` for test database setup
- `db/base.go` — `Base` struct with UUID generation (used in `BeforeCreate` hook)

### Testing Patterns
- Go testing package docs: `testing` — standard `func TestXxx(t *testing.T)` pattern
- GORM testing with SQLite: `gorm.Open(sqlite.Open(":memory:"), ...)` for in-memory DB

</canonical_refs>

<specifics>
## Specific Ideas

### Bug Details
1. **Typo handler:** `DeletePodcasDeleteOnlyPodcasttEpisodesById` (line 162, `controllers/podcast.go`) duplicates `DeletePodcastEpisodesById` (line 151). Identical body. Safe to delete.
2. **removeStartingSlash:** Current logic at `main.go:44-50`:
   ```go
   "removeStartingSlash": func(raw string) string {
       fmt.Println(raw)
       if string(raw[0]) == "/" {
           return raw   // BUG: keeps slash
       }
       return "/" + raw  // BUG: adds slash
   },
   ```
   Fix: remove `fmt.Println`, reverse the branches.

### Test Structure
- `db/db_test.go` — Test DB layer with `:memory:` SQLite. Verify `AddPodcast` creates a record, `GetPodcastById` returns it, `DeletePodcast` removes it.
- `service/podcastService_test.go` — Test service layer. Verify `AddPodcast` downloads RSS and creates episodes, `GetAllPodcasts` returns with stats, `DeletePodcast` cleans up.

</specifics>

<deferred>
## Deferred Ideas

- Comprehensive test coverage for file service, websocket handlers, and controllers — deferred to Phase 3 (Correctness & Concurrency Fixes)
- Testify/assert migration — if Go stdlib proves too verbose in practice, reconsider after this phase
- Mock HTTP server for testing RSS feed parsing — not needed for basic CRUD tests
- Benchmark tests for DB queries — future optimization work

</deferred>

---

*Phase: 02-test-framework-code-quality*
*Context gathered: 2026-05-12 via discuss-phase*

# Phase 3: Correctness & Concurrency Fixes - Context

**Gathered:** 2026-05-12
**Status:** Ready for planning

## Phase Boundary

Fix four correctness and concurrency bugs in the Go podcast manager, then verify them with targeted tests. Building on Phase 2's test harness, this phase addresses real reliability issues: unbounded download concurrency, fragile RSS date parsing, missing database initialization checks, and racy WebSocket state. No new features — only fixes that make the download-and-organize loop reliable.

**In scope:**
- BUG-04: Download concurrency limit is ignored — `sync.WaitGroup` doesn't limit goroutines
- BUG-05: RSS date parsing panics on non-RFC1123 formats (ISO 8601, RFC 3339, custom strings)
- BUG-06: Database initialization failure returns nil DB silently, causing nil pointer dereferences later
- BUG-07: WebSocket connection maps (`activePlayers`, `allConnections`) accessed without mutex
- TEST-04: Verification test for download concurrency limit
- TEST-05: Verification test for date parsing on multiple formats

**Out of scope:**
- Full WebSocket test coverage (beyond data race fix)
- File service refactoring (Phase 4)
- Error handling modernization (Phase 4)
- Comprehensive download integration tests

## Accumulated Context

### From Phase 2 (Test Framework & Code Quality)
- Test harness established: `go test ./...` passes
- `db_test.go` and `service_test.go` created and passing
- SQLite `:memory:` database setup pattern established in test helpers
- GORM v1.26.0 with WAL mode active

### Known Bugs (from Architecture / Codebase Audit)

**BUG-04 — Download Batching (service/fileService.go)**
- Current code: loops through podcast items and launches one goroutine per item with `sync.WaitGroup`. No semaphore or worker pool.
- Expected behavior: only N concurrent downloads at a time (where N is configurable via `CHECK_FREQUENCY` or a hardcoded limit).
- Impact: unbounded goroutines = memory pressure + potential file descriptor exhaustion on large podcast feeds.

**BUG-05 — Date Parsing (service/podcastService.go)**
- Current code: tries `time.Parse(time.RFC1123, date)` only.
- Real-world RSS feeds use: RFC1123Z, RFC1123 `with`, ISO 8601 (`2006-01-02T15:04:05Z07:00`), and custom formats like `Mon, 02 Jan 2006 15:04:05 MST`.
- Impact: `time.Parse` error silently ignored → episodes get `0001-01-01` release dates, breaking "latest episodes" ordering.

**BUG-06 — DB Initialization (db/db.go)**
- Current code: `OpenDB()` returns `(*gorm.DB, error)` but callers in `main.go` check error only with `log.Fatal` if configured. The default path silently ignores the error.
- Impact: if DB open fails (permissions, disk full, corrupted file), `db.DB` is nil → nil pointer panic on first query.

**BUG-07 — WebSocket Races (controllers/websockets.go)**
- Current code: global maps `activePlayers`, `allConnections` accessed without any mutex.
- Impact: `fatal error: concurrent map read and map write` when two clients connect/disconnect simultaneously.
- Fix: add `sync.RWMutex` around map access, or use `sync.Map`.

## Requirements Mapping

| Requirement | Bug/Fix | Test | Target File |
|---|---|---|---|
| BUG-04 | Add buffered channel semaphore to download loop | TEST-04 | service/fileService.go |
| BUG-05 | Try multiple date formats, fallback to zero value with log | TEST-05 | service/podcastService.go |
| BUG-06 | Log.Fatalf on DB init error in main.go | TEST-06 | db/db.go, main.go |
| BUG-07 | Wrap WebSocket map access in mutex | TEST-07 | controllers/websockets.go |

## Decisions Needed

1. **Download limit**: hardcode (e.g., 5) or read from env/config?
2. **Date parsing priority**: which formats to try, in what order?
3. **WebSocket mutex strategy**: `sync.RWMutex` or `sync.Map`?

## Risks

- Date parsing changes could reorder existing episodes in UI — need migration consideration.
- WebSocket mutex adds locking overhead to every message broadcast — measure if concerned.
- DB init failure on existing deployments: `log.Fatalf` will crash the container on startup instead of silently failing later — this is the intended behavior, but verify Docker healthcheck exists.

## Success Criteria Recap

1. Download concurrency limit enforced — verified by test
2. RSS date parsing handles ISO 8601, RFC 3339, custom — verified by test
3. DB init failure aborts app immediately — verified by code review
4. WebSocket maps race-free — verified by `go test -race`
5. All tests pass with `-race` flag

## Dependencies

- Phase 2 complete (test harness in place)
- No new external dependencies expected (stdlib only)

## Files Likely Modified

- `service/fileService.go` — download concurrency
- `service/podcastService.go` — date parsing
- `db/db.go` — DB init error handling
- `controllers/websockets.go` — mutex protection
- New test files: `service/fileService_test.go`, `service/podcastService_date_test.go`, `controllers/websockets_test.go`, `db/db_init_test.go`

## Related Decisions from PROJECT.md

- [Arch]: Single-process, multi-goroutine — download uses WaitGroup, cron runs in background
- [Arch]: WebSocket state held in package-level maps — not thread-safe (to be fixed)
- [Tech]: SQLite via GORM — DB init must succeed or app must not start

## Notes

- This phase is the last "fixing bugs" phase before Phase 4 (Error Handling Modernization). After this, the app should be functionally correct and safe under concurrent use.
- The `go test -race` flag is the primary verification tool for this phase.
- All fixes must be backwards-compatible with existing SQLite databases and downloaded files.

## Created

2026-05-12 — Phase 3 added via `/gsd-phase`

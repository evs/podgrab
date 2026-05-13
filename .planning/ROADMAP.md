# Roadmap: Podgrab

## Overview

Stabilize the existing Go podcast manager: upgrade the dependency foundation, fix bugs with tests to verify them, then modernize error handling. Each phase delivers a coherent, verifiable capability that the next phase depends on — the download-and-organize loop must never break.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3, 4): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Dependency Upgrades** - Upgrade Go to 1.24+, replace abandoned libs, update Docker
- [x] **Phase 2: Test Framework & Code Quality** - Set up test harness, fix trivial code bugs
- [x] **Phase 3: Correctness & Concurrency Fixes** - Fix download batching, date parsing, DB init, WebSocket races
- [x] **Phase 4: Error Handling Modernization** - Structured logging, error propagation, remove panics

## Phase Details

### Phase 1: Dependency Upgrades
**Goal**: App compiles and runs on modern Go and dependency stack
**Mode**: mvp
**Depends on**: Nothing (first phase)
**Requirements**: DEPS-01, DEPS-02, DEPS-03, DEPS-04, DEPS-05, DEPS-06, DEPS-07, DEPS-08, DEPS-09, DEPS-10
**Success Criteria** (what must be TRUE):
  1. `go build` succeeds with Go 1.24+ and all upgraded dependencies
  2. App starts and serves HTTP requests with upgraded GORM and Gin
  3. Docker container builds and runs with updated Dockerfile and docker-compose.yml
  4. PodcastIndex API calls authenticate via environment variables (`PODCASTINDEX_KEY`, `PODCASTINDEX_SECRET`), not hardcoded credentials
  5. No deprecated `io/ioutil` calls remain in the codebase
**Plans**: 4 plans (all complete)
- [x] 01-01-PLAN.md — Go 1.24 upgrade + ioutil replacement
- [x] 01-02-PLAN.md — Remove jwt-go + swap abandoned libraries (uuid, cron, websocket)
- [x] 01-03-PLAN.md — Upgrade Gin + remaining deps + GORM v1.26 migration
- [x] 01-04-PLAN.md — Extract credentials + update Docker config
	
### Phase 2: Test Framework & Code Quality
**Goal**: Tests can be written and run; trivial code bugs are fixed and verified
**Mode**: mvp
**Depends on**: Phase 1
**Requirements**: TEST-01, BUG-01, BUG-02, BUG-03, TEST-02, TEST-03
**Success Criteria** (what must be TRUE):
  - [x] `go test ./...` runs successfully with a test harness in place
  - [x] Podcast delete handler has a single correctly-named function (no duplicate typo handler)
  - [x] `removeStartingSlash` template function removes leading slashes instead of adding them
  - [x] No debug `fmt.Println` remains in `main.go` for `removeStartingSlash`
  - [x] Service and DB layer tests pass for core podcast CRUD operations
**Plans**: 4 plans (all complete)
- [x] 02-01-PLAN.md — test harness + SQLite :memory: harness + first service test
- [x] 02-02-PLAN.md — remove duplicate typo handler `DeletePodcasDeleteOnlyPodcasttEpisodesById`
- [x] 02-03-PLAN.md — fix `removeStartingSlash` logic and remove debug prints
- [x] 02-04-PLAN.md — DB + Service layer CRUD tests

### Phase 3: Correctness & Concurrency Fixes
**Goal**: Download scheduling works reliably, date parsing handles real RSS feeds, crashes are caught early
**Mode**: mvp
**Depends on**: Phase 2
**Requirements**: BUG-04, BUG-05, BUG-06, BUG-07, TEST-04, TEST-05
**Success Criteria** (what must be TRUE):
  1. [x] Download concurrency limit is actually enforced — only N episodes download in parallel
  2. [x] RSS date parsing handles ISO 8601, RFC 3339, and common podcast date formats
  3. [x] App exits with fatal error if database initialization fails (no silent nil DB)
  4. [x] WebSocket connections don't cause data races under concurrent access
  5. [x] Download concurrency and date parsing have passing verification tests
**Plans**: 1 plan (complete), 6 tasks
- [x] 03-PLAN.md — BUG-05 parseRSSDate, BUG-04 semaphore download, BUG-06 DB init error propagation, BUG-07 WebSocket RWMutex, TEST-04/05 verification tests

### Phase 4: Error Handling Modernization
**Goal**: Errors are visible, structured, and properly surfaced instead of silently swallowed
**Mode**: mvp
**Depends on**: Phase 3
**Requirements**: ERR-01, ERR-02, ERR-03, ERR-04, ERR-05
**Success Criteria** (what must be TRUE):
  1. [x] All error output uses structured zap logging instead of `fmt.Println`/`log.Println`
  2. [x] Service/DB errors are returned to HTTP handlers instead of silently ignored
  3. [x] HTTP handlers return proper 4xx/5xx status codes on errors instead of empty or wrong responses
  4. [x] No panic-on-error patterns remain in `service/fileService.go`
  5. [x] Error handling conventions documented for future development
**Plans**: 1 plan (complete)
- [x] 04-PLAN.md — Replace fmt.Println with structured zap logging, propagate errors, fix HTTP status codes, remove checkError panic

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Dependency Upgrades | 4/4 | ✓ Complete | 2026-05-12 |
| 2. Test Framework & Code Quality | 4/4 | ✓ Complete | 2026-05-12 |
| 3. Correctness & Concurrency Fixes | 1/1 | ✓ Complete | 2026-05-13 |
| 4. Error Handling Modernization | 1/1 | ✓ Complete | 2026-05-13 |

**Milestone v1.26: 100% COMPLETE**

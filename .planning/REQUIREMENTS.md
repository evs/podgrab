# Requirements: Podgrab

**Defined:** 2026-05-12
**Core Value:** Podcast episodes are automatically downloaded and available — that download-and-organize loop must never break

## v1 Requirements

Requirements for stabilization milestone. Each maps to roadmap phases.

### Dependency Upgrades

- [x] **DEPS-01**: Go version upgraded from 1.15 to 1.24+ (go.mod updated, code compiles and runs)
- [x] **DEPS-02**: GORM upgraded from v1.20.2 to v1.26+ (migration accounts for silent runtime behavior changes)
- [x] **DEPS-03**: `satori/go.uuid` replaced with `google/uuid` (latest stable)
- [x] **DEPS-04**: `dgrijalva/jwt-go` replaced with `golang-jwt/jwt/v5` (or removed if unused)
- [x] **DEPS-05**: `jasonlvhit/gocron` replaced with `robfig/cron/v3`
- [x] **DEPS-06**: Gin upgraded to latest stable (v1.12+)
- [x] **DEPS-07**: All other dependencies upgraded to current versions via `go get -u`
- [x] **DEPS-08**: Deprecated `io/ioutil` usage replaced with `io.ReadAll`, `os.WriteFile`, etc.
- [x] **DEPS-09**: Hardcoded PodcastIndex API credentials moved to environment variables (`PODCASTINDEX_KEY`, `PODCASTINDEX_SECRET`)
- [x] **DEPS-10**: Dockerfile and docker-compose.yml updated for new Go version and build requirements

### Bug Fixes

- [ ] **BUG-01**: Duplicate typo handler `DeletePodcasDeleteOnlyPodcasttEpisodesById` removed or corrected in `controllers/podcast.go`
- [ ] **BUG-02**: `removeStartingSlash` template function fixed — actually removes leading slash instead of adding one
- [ ] **BUG-03**: Debug `fmt.Println` in `removeStartingSlash` removed from `main.go`
- [x] **BUG-04**: `DownloadMissingEpisodes` concurrency batching fixed — `wg.Wait()` logic corrected so concurrency limit is actually enforced
- [x] **BUG-05**: Date parsing expanded to handle ISO 8601, RFC 3339, and other common RSS date formats beyond current 5 hardcoded formats
- [x] **BUG-06**: DB init error causes fatal exit instead of continuing with nil DB (fixes `statuse` typo too)
- [x] **BUG-07**: WebSocket state maps (`activePlayers`, `allConnections`) protected with `sync.RWMutex` or `sync.Map` to prevent data races

### Error Handling

- [x] **ERR-01**: `fmt.Println`/`log.Println` error logging replaced with structured `zap.SugaredLogger` calls throughout codebase
- [x] **ERR-02**: Errors from service/db calls propagated to HTTP handler layer instead of silently ignored
- [x] **ERR-03**: HTTP handlers return appropriate status codes (4xx/5xx) when errors occur instead of empty or wrong responses
- [x] **ERR-04**: `checkError` panic-on-error pattern in `service/fileService.go` replaced with proper error return
- [x] **ERR-05**: Consistent error handling pattern established and documented for future development

### Test Coverage

- [ ] **TEST-01**: Go test framework set up (`testing` package, test helpers, test runner)
- [ ] **TEST-02**: Service layer unit tests for core podcast operations (add, refresh, download, delete)
- [ ] **TEST-03**: DB layer tests for CRUD operations and migrations using test SQLite database
- [ ] **TEST-04**: Date parsing tests covering all supported RSS date formats
- [ ] **TEST-05**: Download concurrency tests verifying batching limits work correctly

## v2 Requirements

Deferred to future milestone after stabilization assessment.

### REST API Layer

- **API-01**: `/api/v1/*` route group with versioned endpoints
- **API-02**: DTO package for request/response models separate from internal DB models
- **API-03**: All controller `db.*` direct calls routed through service layer
- **API-04**: API authentication via existing HTTP Basic Auth or token-based auth

### React Frontend

- **UI-01**: Vite + React 19 + React Router 7 scaffold served alongside existing SSR UI
- **UI-02**: Responsive design (mobile-friendly)
- **UI-03**: Dark mode with CSS `prefers-color-scheme` + manual toggle
- **UI-04**: Persistent audio player bar (sticky bottom)
- **UI-05**: Playback speed control
- **UI-06**: Episode play progress tracking persisted to DB
- **UI-07**: Played/unplayed episode indicators
- **UI-08**: Episode search/filter by title, date, download status, played status
- **UI-09**: Episode sorting controls (newest, oldest, alphabetical)
- **UI-10**: Download status indicators (queued/downloading/complete/error)
- **UI-11**: Episode description/show notes rendering
- **UI-12**: Keyboard shortcuts (play/pause, seek, speed, volume)

### Differentiators

- **DIFF-01**: Per-podcast auto-download rules (latest N only, manual-only, etc.)
- **DIFF-02**: Episode timeline / "what's new" dashboard across all subscriptions
- **DIFF-03**: RSS feed re-scan on demand per podcast
- **DIFF-04**: Batch episode actions (select multiple → mark played, delete, download)
- **DIFF-05**: Disk space management / auto-cleanup of played episodes
- **DIFF-06**: PWA support
- **DIFF-07**: Subscription health / feed error visibility

## Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-user / RBAC | Single-user tool, basic auth is adequate; Audiobookshelf for multi-user |
| Audiobook management | Not a media server; Audiobookshelf exists for this |
| Native mobile app | PWA is 80% of the benefit at 10% of the cost |
| Social features | Not core to download-and-organize value |
| Podcast streaming | Podgrab is download-first; streaming explicitly out of scope |
| AI features | Orthogonal to core value, requires external resources |
| HLS streaming | Rarely used by podcasts, adds player complexity |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DEPS-01 | Phase 1 | Complete |
| DEPS-02 | Phase 1 | Complete |
| DEPS-03 | Phase 1 | Complete |
| DEPS-04 | Phase 1 | Complete |
| DEPS-05 | Phase 1 | Complete |
| DEPS-06 | Phase 1 | Complete |
| DEPS-07 | Phase 1 | Complete |
| DEPS-08 | Phase 1 | Complete |
| DEPS-09 | Phase 1 | Complete |
| DEPS-10 | Phase 1 | Complete |
| TEST-01 | Phase 2 | Complete |
| BUG-01 | Phase 2 | Complete |
| BUG-02 | Phase 2 | Complete |
| BUG-03 | Phase 2 | Complete |
| TEST-02 | Phase 2 | Complete |
| TEST-03 | Phase 2 | Complete |
| BUG-04 | Phase 3 | Complete |
| BUG-05 | Phase 3 | Complete |
| BUG-06 | Phase 3 | Complete |
| BUG-07 | Phase 3 | Complete |
| TEST-04 | Phase 3 | Complete |
| TEST-05 | Phase 3 | Complete |
| ERR-01 | Phase 4 | Complete |
| ERR-02 | Phase 4 | Complete |
| ERR-03 | Phase 4 | Complete |
| ERR-04 | Phase 4 | Complete |
| ERR-05 | Phase 4 | Complete |

**Coverage:**
- v1 requirements: 27 total
- Mapped to phases: 27
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-12*
*Last updated: 2026-05-12 after questioning revision — stabilize-first scope*
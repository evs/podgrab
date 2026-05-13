# Phase 4: Error Handling Modernization

**Plan:** 04-PLAN.md  
**Date:** 2026-05-13  
**Mode:** MVP  

---

## Summary

Modernize error handling across the Go codebase by replacing `fmt.Println`/`log.Println` with structured `zap` logging, propagating errors to HTTP handlers, returning correct HTTP status codes, removing the `checkError` panic-on-error pattern, and documenting conventions. 3 waves, ~30% context usage target.

---

## Requirements Mapping

| Requirement | Plan Coverage |
|-------------|---------------|
| ERR-01 | Wave 1 tasks 1–6, Wave 2 tasks 1–4 |
| ERR-02 | Wave 1 tasks 2, 4; Wave 2 tasks 2, 4 |
| ERR-03 | Wave 2 tasks 1–4 |
| ERR-04 | Wave 1 task 5 |
| ERR-05 | Wave 3 task 1 |

---

## Decisions

1. Use existing `zap.SugaredLogger` (`service.Logger`) everywhere in service layer.
2. Controllers will import `go.uber.org/zap` and create their own `*zap.SugaredLogger` instance (or reuse via exported getter). Simplest: initialize `var Logger *zap.SugaredLogger` in `controllers` package.
3. HTTP status mapping: user input error → 400, not found → 404, internal/db/file → 500.
4. `checkError` removed; `getFileName` → `(string, error)`; single call site updated.

---

## Wave 1: Service Layer

Wave 1 focuses on `service/*.go` files. Replace all `fmt.Println`/`log.Println` with structured logging; propagate errors up; replace panic with error return.

### Task 1.1: Replace `fmt.Println`/`log.Println` in `service/podcastService.go`
- **files**: `service/podcastService.go`
- **action**:
  1. Search file for all `fmt.Println(...)` and `fmt.Println(err.Error())` calls (14 active occurrences; ignore commented-out ones).
  2. Replace each with the appropriate `Logger` call:
     - Errors → `Logger.Errorw("description", zap.Error(err))`
     - Debug info → `Logger.Infow("description", "key", value)` or `Logger.Debugw(...)`
  3. Ensure no `fmt.Println` remains except commented-out debug lines.
- **verify**:
  ```bash
  grep -n "fmt.Println" service/podcastService.go | grep -v "//fmt.Println"
  # must be empty
  ```
- **done**: Yes

### Task 1.2: Propagate errors in `service/podcastService.go` where currently swallowed
- **files**: `service/podcastService.go`
- **action**:
  1. Find every function that currently calls `fmt.Println(err)` and then does `return nil` or drops the error.
  2. Remove the print; return `err` (or wrap it) so the controller receives it.
  3. Example: if a DB query fails inside `RefreshEpisodes`, return the error so the cron loop or controller can react.
- **verify**:
  ```bash
  grep -n "fmt.Println" service/podcastService.go | grep -v "//fmt.Println"
  # empty AND go vet ./service/... shows no dropped error values
  ```
- **done**: Yes

### Task 1.3: Replace `fmt.Println` in `service/itunesService.go`
- **files**: `service/itunesService.go`
- **action**:
  1. Replace lines 45, 53 `fmt.Println(...)` with `Logger.Errorw` or `Logger.Warnw`.
  2. Ensure `Logger` variable is accessible (import `go.uber.org/zap` if needed, or use `service.Logger`).
- **verify**:
  ```bash
  grep -n "fmt.Println" service/itunesService.go
  # must be empty
  ```
- **done**: Yes

### Task 1.4: Replace `fmt.Println` in `service/fileService.go`
- **files**: `service/fileService.go`
- **action**:
  1. Replace the 2 `fmt.Println` calls with structured logging.
  2. Ensure file-path logging is safe (no leaking of full absolute paths to external log sinks if a concern; in this self-hosted app, `Debugw` is fine).
- **verify**:
  ```bash
  grep -n "fmt.Println" service/fileService.go
  # must be empty
  ```
- **done**: Yes

### Task 1.5: Remove `checkError` panic-on-error; refactor `getFileName`
- **files**: `service/fileService.go`
- **action**:
  1. Delete `checkError` function (lines ~404–408).
  2. Change `getFileName(link string) string` → `getFileName(link string) (string, error)`.
  3. Replace `checkError(err)` inside `getFileName` with `if err != nil { return "", err }`.
  4. Update the single call site (~line 386, inside `Download` or nearby) to capture the error and propagate/return it.
  5. Ensure any intermediate callers (e.g., `downloadImageLocally`, `createFolder` if involved) are updated to handle the new error return.
- **verify**:
  ```bash
  grep -n "checkError" service/fileService.go
  # must be empty
  grep -n "func getFileName" service/fileService.go
  # signature must be: func getFileName(link string) (string, error)
  ```
- **done**: Yes

### Task 1.6: Build and lint Wave 1 changes
- **files**: `service/*.go`
- **action**:
  1. Run `go build ./service/...` to confirm no compile errors.
  2. Run `go vet ./service/...`.
  3. Run `go test ./service/...` (or `go test ./...` if tests exist).
- **verify**:
  ```bash
  go build ./service/...   # exit 0
  go vet ./service/...     # exit 0
  go test ./...          # passes
  ```
- **done**: Yes

---

## Wave 2: Controller Layer

Wave 2 focuses on `controllers/*.go`. Replace all `fmt.Println(err)` with logging + proper JSON error responses. Map errors to correct HTTP status codes.

### Task 2.1: Replace `fmt.Println` and add structured logging in `controllers/podcast.go`
- **files**: `controllers/podcast.go`
- **action**:
  1. Add `var Logger *zap.SugaredLogger` and init function at top of file (or import from `service`).
  2. Find all 7 `fmt.Println(err)` / `fmt.Println(err.Error())` occurrences (lines 92, 171, 184, 196, 220, 544, 632 per research).
  3. Replace each:
     - Log: `Logger.Errorw("handler error", zap.Error(err))`
     - Return JSON error response: `c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})` (or 400/404 as appropriate).
- **verify**:
  ```bash
  grep -n "fmt.Println" controllers/podcast.go
  # must be empty
  ```
- **done**: Yes

### Task 2.2: Return correct HTTP status codes in `controllers/podcast.go`
- **files**: `controllers/podcast.go`
- **action**:
  1. For each handler, review error paths:
     - `gorm.ErrRecordNotFound` → `c.JSON(404, gin.H{"error": "not found"})`
     - Custom errors like `*model.PodcastAlreadyExistsError` → `400`
     - Invalid query/body binding → `400`
     - All other internal errors → `500`
  2. Ensure success paths still return `200` / `201` as before.
- **verify**:
  ```bash
  go build ./controllers/...  # exit 0
  # Manual spot-check: every c.JSON call in podcast.go must have an HTTP status literal:
  grep -n "c.JSON(" controllers/podcast.go | grep -v "http.Status" | grep -v "200\|201"
  # should be empty (or only lines already using a status variable)
  ```
- **done**: Yes

### Task 2.3: Replace `fmt.Println` in `controllers/pages.go`
- **files**: `controllers/pages.go`
- **action**:
  1. Replace lines 174, 387 `fmt.Println(err.Error())` with structured logging + appropriate JSON or template error handling.
  2. For server-side rendered pages, if an error occurs during page load, either:
     - Log it and render an error page/message, OR
     - Return `c.String(500, err.Error())` (acceptable for page handlers in this phase).
- **verify**:
  ```bash
  grep -n "fmt.Println" controllers/pages.go
  # must be empty
  ```
- **done**: Yes

### Task 2.4: Replace `fmt.Println` in `controllers/websockets.go`
- **files**: `controllers/websockets.go`
- **action**:
  1. Replace lines 40, 131 `fmt.Println(...)` with structured logging.
  2. WebSocket handlers should not crash the server on parse errors; log and drop malformed messages.
  3. If `Logger` not already present, initialize alongside existing `activePlayers` map (or reuse from service).
- **verify**:
  ```bash
  grep -n "fmt.Println" controllers/websockets.go
  # must be empty
  ```
- **done**: Yes

---

## Wave 3: Verification & Documentation

### Task 3.1: Final acceptance verification
- **files**: entire project
- **action**:
  1. Run the full acceptance criteria suite:
     ```bash
     grep -r "fmt.Println(err)" service/     # ERR-01
     grep -r "fmt.Println(err)" controllers/ # ERR-01 / ERR-02
     grep -n "checkError" service/fileService.go  # ERR-04
     go test ./...                           # build + test
     go build ./...                          # ERR-05 indirectly
     ```
  2. If any grep matches, fix them before marking this task done.
- **verify**:
  - All three `grep` commands return empty.
  - `go test ./...` exits 0.
  - `go build ./...` exits 0.
- **done**: Yes

### Task 3.2: Document error handling conventions
- **files**: `.planning/phases/04-error-handling-modernization/04-PLAN.md` (self-reference), inline comments in code
- **action**:
  1. Add a short header comment block at the top of `service/podcastService.go` (or a new `doc.go` in `controllers/`) summarizing the convention:
     ```
     // Error handling convention for this package:
     // - Service errors are logged with Logger.Errorw and returned to callers.
     // - Controllers map errors to HTTP status codes:
     //     user input errors      → 400 Bad Request
     //     not found (DB)         → 404 Not Found
     //     internal/service errors → 500 Internal Server Error
     // - Never swallow errors with fmt.Println; never panic on recoverable errors.
     ```
  2. Ensure all modified functions have at least one inline comment where the error handling pattern changed.
- **verify**:
  ```bash
  grep -A5 "Error handling convention" service/podcastService.go controllers/podcast.go
  # at least one occurrence found
  ```
- **done**: Yes

---

## Acceptance Criteria

| # | Criterion | Verification Command |
|---|-----------|----------------------|
| 1 | No `fmt.Println(err)` in `service/` | `grep -r "fmt.Println(err)" service/` → empty |
| 2 | No `fmt.Println(err)` in `controllers/` | `grep -r "fmt.Println(err)" controllers/` → empty |
| 3 | `checkError` removed from `service/fileService.go` | `grep -n "checkError" service/fileService.go` → empty |
| 4 | Tests pass | `go test ./...` → exit 0 |
| 5 | Build passes | `go build ./...` → exit 0 |
| 6 | All handlers return `c.JSON` for success and error paths | `grep -n "c.JSON(" controllers/*.go` → every handler has `c.JSON` on both success and at least one error branch |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Changing many files at once introduces subtle bugs | Do it in waves (service first, then controllers); build/test after each wave. |
| Zap logger not initialized in tests | If `service.Logger` is nil in unit tests, wrap with `if Logger != nil` or initialize in a test setup file. |
| `getFileName` error propagation breaks download flow | Test `go test ./...` after task 1.5 to confirm callers handle the new error return. |
| Controllers returning 500 for benign “already exists” | Check for custom errors (`PodcastAlreadyExistsError`, etc.) and map to 400. |

---

## Estimates

| Wave | Estimated Duration |
|------|--------------------|
| Wave 1 (Service) | 6 tasks × 3 min = 18 min |
| Wave 2 (Controllers) | 4 tasks × 4 min = 16 min |
| Wave 3 (Verify + Docs) | 2 tasks × 3 min = 6 min |
| **Total** | **~40 min** |

---

## Plan Meta

- **Phase**: 04-error-handling-modernization
- **Plan**: 04-PLAN.md
- **Created**: 2026-05-13
- **Mode**: MVP
- **Waves**: 3
- **Tasks**: 10

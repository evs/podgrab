---
phase: 04-error-handling-modernization
name: Error Handling Modernization
gathered: 2026-05-13
status: Ready for planning
---

# Phase 4: Error Handling Modernization — Context

## Task Boundary

Modernize the entire Go podcast manager's error handling:
1. Replace `fmt.Println`/`log.Println` with structured `zap.SugaredLogger`.
2. Stop swallowing errors silently in controllers — always return them to the HTTP client.
3. Fix HTTP handlers to return correct 4xx/5xx status codes.
4. Remove `checkError` panic-on-error in `service/fileService.go`.
5. Document error handling conventions.

## Scope

**In scope:**
- `service/podcastService.go` (14 `fmt.Println` occurrences)
- `service/itunesService.go` (2 `fmt.Println` occurrences)  
- `service/fileService.go` (`checkError` panic, 2 `fmt.Println`)
- `controllers/podcast.go` (7 `fmt.Println(err)` occurrences)
- `controllers/pages.go` (2 `fmt.Println` occurrences)
- `controllers/websockets.go` (2 `fmt.Println` occurrences)

**Out of scope:**
- `main.go` `log.Println` (startup diagnostics, acceptable)
- Commented-out `//fmt.Println` (leave as-is)
- `service/naturaltime.go` (commented debug only)
- No database schema changes.
- No new behavioral features.

## Implementation Decisions

### 1. Which logger to use?
- **Decision:** Use the existing `zap.SugaredLogger` (variable `Logger`) already initialized in `service/podcastService.go`.
- **Rationale:** Already present, configured with production settings; no new dependency needed.

### 2. Where to add Logger in controllers?
- **Decision:** Create a package-level `Logger` in `controllers` by importing `go.uber.org/zap` and calling `zap.NewProduction()` (or expose from `service` package — re-expose is acceptable).
- **Rationale:** Controllers should log, too; reusing `service.Logger` via a getter or moving to a shared logger package is cleaner. Phase 4 is MVP; simple is best.

### 3. What to do with `checkError` panic?
- **Decision:** Remove `checkError()`. Change `getFileName()` to return `(string, error)`. Update its single call site in `fileService.go:386`.
- **Rationale:** Panic-on-error is exactly the anti-pattern this phase targets. Normal error propagation is correct.

### 4. How to handle errors in HTTP handlers that currently print + ignore?
- **Decision:**
  - If the error comes from a user-provided input (e.g., bad podcast URL, invalid ID) → return `400 Bad Request` with `gin.H{"error": err.Error()}`.
  - If the error is internal (e.g., DB failure, file I/O) → return `500 Internal Server Error`.
  - If the error is a "not found" (`gorm.ErrRecordNotFound`) → return `404 Not Found`.
  - **Never** leave an error unhandled (no bare `fmt.Println(err)` + `c.JSON(200, ...)`).

### 5. How to organize tasks?
- **Decision:** Wave-based execution:
  - Wave 1: Service layer (`service/*.go`) — replace `fmt.Println` with structured logging, remove `checkError`.
  - Wave 2: Controller layer (`controllers/*.go`) — fix error propagation and HTTP status codes.
  - Wave 3: Documentation + final `go test` verification.

### 6. Logging level mapping
- **Decision:**
  - Error conditions (`err != nil`) → `Logger.Errorw("...", "error", err)`
  - Informational diagnostics ("Processing N episodes") → `Logger.Infow("...", "count", N)`
  - Warnings (missing env var, external API failure) → `Logger.Warnw("...", "error", err)`

## Specific Ideas

- In `service/podcastService.go`, the existing `Logger` is already a sugared logger, so direct use is `Logger.Errorw("message", "error", err, "field", value)`.
- For controllers: create a simple helper `func respondWithError(c *gin.Context, status int, err error)` or just inline the `c.JSON(...)` pattern — Gin is already imported.
- The `getFileName` caller is at `service/fileService.go:386`. Search for `getFileName(` to find all call sites.

## Canonical References

- [Podgrab Architecture: Error Handling section](../ARCHITECTURE.md#error-handling)
- [Podgrab Conventions: Logging section](../CONVENTIONS.md#logging)
- [Podgrab Requirements: ERR-01 through ERR-05](../REQUIREMENTS.md)

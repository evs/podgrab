# Testing Patterns

**Analysis Date:** 2026-05-11

## Test Framework

**Runner:**
- Not applicable — **no test framework is configured or used in this project**
- Go's built-in `testing` package would be the standard, but no `*_test.go` files exist
- No test runner configuration files found (`Makefile`, `justfile`, test scripts)

**Assertion Library:**
- None configured

**Run Commands:**
```bash
go test ./...              # Would run tests (none exist)
go test -v ./...          # Verbose mode (none exist)
go test -cover ./...       # Coverage (none exist)
```

## Test File Organization

**Location:**
- No test files exist anywhere in the project
- No `*_test.go` files found via glob search
- No `test/` or `tests/` directories

**Naming:**
- Go convention would be `*_test.go` alongside source files
- Package-level tests: `podcastService_test.go` in `service/`
- No Black-box test packages observed

**Structure:**
```
[Not applicable — no tests exist]
```

## Test Structure

**Suite Organization:**
- No test suites defined
- Go convention would use `func TestXxx(t *testing.T)` per function

**Patterns:**
- No setup/teardown patterns
- No table-driven test patterns
- No test helpers

## Mocking

**Framework:** None

**Patterns:**
- No mocking frameworks configured
- No interfaces defined for testability (DB operations use concrete `*gorm.DB` via global `db.DB`)
- Service functions call `db.*` package functions directly — no dependency injection
- HTTP client creation in `service/fileService.go:336` uses `http.Client{}` directly with no interface, making HTTP mocking difficult
- `makeQuery` (`service/podcastService.go:710`) constructs HTTP requests directly — no HTTP client interface

**What to Mock (once tests exist):**
- Database layer (`db.DB` global) — would need interface wrapper around GORM
- HTTP client calls (`makeQuery`, `Download`, `DownloadImage`, etc.)
- File system operations (`os.Create`, `os.Stat`, `os.MkdirAll`, `os.Remove`, `os.Chown`)
- Time-dependent functions (cron scheduling, lock expiration)

**What NOT to Mock:**
- Model structs (pure data)
- String manipulation functions (`sanitize.Name`, `sanitize.Path`)
- `NatualTime` (pure function with deterministic input/output)

## Fixtures and Factories

**Test Data:**
- No fixture files found
- No factory functions for creating test entities
- Seed data would need to be created programmatically via GORM

**Location:**
- No `testdata/` directories
- No `fixtures/` directories

## Coverage

**Requirements:** None enforced — no CI test step, no coverage targets

**View Coverage:**
```bash
go test -coverprofile=coverage.out ./...    # Would generate coverage (no tests)
go tool cover -html=coverage.out             # Would view HTML report
```

## Test Types

**Unit Tests:**
- Not present
- Candidates: `service/naturaltime.go` (pure functions), `internal/sanitize/sanitize.go` (pure functions), `model/errors.go` (simple error types)

**Integration Tests:**
- Not present
- Would need: database layer tests, HTTP handler tests via `httptest`

**E2E Tests:**
- Not used
- No browser automation or API test suites

## Common Patterns

**Async Testing:**
```go
// Not applicable — no tests exist
// Goroutines used in service layer (go service.RefreshEpisodes(), go service.DownloadSingleEpisode)
// Testing concurrent code would require sync primitives or channel-based coordination
```

**Error Testing:**
```go
// Not applicable — no tests exist
// Custom error types in model/errors.go could be tested:
//   func TestPodcastAlreadyExistsError(t *testing.T) {
//       err := &PodcastAlreadyExistsError{Url: "http://example.com"}
//       if err.Error() == "" {
//           t.Error("Error() should not be empty")
//       }
//   }
```

## CI Pipeline

**Current State:**
- `.github/workflows/hub.yml` — Docker build & push only, no test step
- No linting step
- No test execution in CI
- Triggers on push to `master` only

**Pipeline Steps:**
1. Checkout
2. Set up QEMU
3. Set up Docker Buildx
4. Cache setup
5. Docker Hub login
6. GitHub Container Registry login
7. Build & push multi-platform Docker image

## Testing Recommendations

Since there are currently zero tests, here's how to bootstrap testing:

**Priority 1 — Pure Function Tests (no mocking needed):**
- `service/naturaltime.go`: `NatualTime`, `pastNaturalTime`, `futureNaturalTime` — pure time calculations
- `internal/sanitize/sanitize.go`: `Name`, `Path`, `HTML`, `Accents` — pure string transformations
- `model/errors.go`: Custom error type `Error()` methods
- `model/queryModels.go`: `EpisodesFilter.VerifyPaginationValues()`, `EpisodesFilter.SetCounts()`

**Priority 2 — Controller Handler Tests:**
- Use `net/http/httptest` with `gin.CreateTestContext`
- Mock `db.DB` via GORM's `gorm.DB` interface or use SQLite in-memory for test DB
- Test request binding, response codes, error handling

**Priority 3 — Service Layer Tests:**
- Introduce interfaces for DB operations to enable mocking
- Test `AddPodcast`, `AddTag`, `RefreshEpisodes`, download logic
- Mock HTTP calls via `net/http/httptest.Server`

**Test DB Strategy:**
- `gorm.io/driver/sqlite` already in use — use in-memory SQLite (`:memory:`) for tests
- Create `db.TestInit()` that returns an in-memory GORM instance
- Run migrations on test DB before each test suite

---

*Testing analysis: 2026-05-11*
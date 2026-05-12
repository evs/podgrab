# Phase 2 Research: Test Framework & Code Quality

**Phase:** 2 — Test Framework & Code Quality
**Researched:** 2026-05-12
**Go Version:** 1.25.0 (after Phase 1 upgrade)

---

## 1. Codebase State After Phase 1

### Zero Test Files
No `*_test.go` files exist anywhere in the project. This is a greenfield test setup.

### Package Layout
```
controllers/     # Gin HTTP handlers (3 files, 647+ lines each)
db/              # GORM models + CRUD functions (4 files)
internal/sanitize/  # File name sanitization
model/           # DTOs + error types
service/         # Business logic (5 files, 817 lines in podcastService.go)
main.go          # Router, middleware, template funcs
```

### Key Dependencies (Post-Phase 1)
- `gorm.io/gorm v1.26.0`
- `gorm.io/driver/sqlite v1.5.7`
- `github.com/gin-gonic/gin v1.12.0`
- `github.com/google/uuid v1.6.0`
- `go.uber.org/zap v1.28.0`

### Global Mutable State
- `var DB *gorm.DB` in `db/db.go:14` — global DB handle
- `var Logger *zap.SugaredLogger` in `service/podcastService.go:24`
- Template `funcMap` in `main.go:35`

---

## 2. Testing Strategy

### Decision: Stdlib-Only (No testify)
Per `02-CONTEXT.md` D-01, use `testing` package only. This means:
- No `assert.Equal()` — use `if got != want { t.Errorf(...) }`
- No `require.NoError()` — use `if err != nil { t.Fatalf(...) }` for setup, `t.Errorf` for assertions
- No `testify/mock` — build simple interfaces and manual mocks

### 2.1 GORM + SQLite `:memory:` Setup

Go 1.25 + GORM v1.26 supports SQLite in-memory cleanly:

```go
// In any *_test.go file
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("open test db: %v", err)
    }
    // Migrate schema
    db.AutoMigrate(&db.Podcast{}, &db.PodcastItem{}, &db.Setting{}, ...)
    return db
}
```

**Key insight:** The `:memory:` DSN with `?cache=shared` is required so the same in-memory DB persists across connection opens (GORM may open new connections internally). Without it, AutoMigrate and queries hit different in-memory DBs.

**Alternative:** Use `test-uuid.db` file-based approach if shared cache causes issues — but `:memory:` is faster and cleaner.

### 2.2 TestMain for Shared Setup

For DB layer tests that need consistent schema across all subtests:

```go
func TestMain(m *testing.M) {
    // One-time schema init
    code := m.Run()
    // No cleanup needed for :memory:
    os.Exit(code)
}
```

For per-test isolation (preferred), each test calls `setupTestDB(t)` and inserts its own fixtures. This avoids test ordering bugs.

### 2.3 Mocking External Services

Podgrab calls external APIs via direct function calls (`service.FetchURL`, `service.SearchPodcasts`). For unit testing, the cleanest Go-idiomatic approach:

**Option A: Interface + Dependency Injection (Recommended for Phase 2)**
Introduce small interfaces for the service layer functions we want to test without real HTTP calls:

```go
// In service layer
var httpFetcher HTTPFetcher = defaultFetcher{}

type HTTPFetcher interface {
    Fetch(url string) ([]byte, error)
}

type defaultFetcher struct{}

func (d defaultFetcher) Fetch(url string) ([]byte, error) {
    return makeQuery(url)
}
```

**BUT** this requires significant refactoring. For Phase 2 scope:

**Option B: Skip HTTP mocking, test DB + business logic only**
Focus tests on the DB layer (`db/dbfunctions.go`) and service functions that don't hit the network. The three bug fixes are all in code that doesn't need HTTP mocking.

**For Phase 2, we use Option B** — test:
- DB CRUD functions (no external calls)
- Template functions (pure functions)
- Bug fixes (remove typo handler, fix `removeStartingSlash`)

### 2.4 Gin Handler Testing

Gin provides `httptest` support natively:

```go
import (
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
)

func TestDeletePodcastById(t *testing.T) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{{Key: "id", Value: someUUID}}

    // Need to set up DB first
    db.Init() // or setupTestDB
    controllers.DeletePodcastById(c)

    if w.Code != http.StatusNoContent {
        t.Errorf("DELETE /podcasts/:id = %d, want %d", w.Code, http.StatusNoContent)
    }
}
```

**Challenge:** Controllers depend on `db.DB` global. Tests must initialize it before calling handlers. With `db.Init()` using file path, tests would write to disk. Better approach:

Override `db.DB` in tests:

```go
func setupControllerTest(t *testing.T) {
    testDB := setupTestDB(t)
    db.DB = testDB // override global
    t.Cleanup(func() { db.DB = nil })
}
```

This is pragmatic for a codebase with global state. The alternative (injecting DB into every handler) is Phase 4+ work.

---

## 3. Known Bugs — Detailed Findings

### 3.1 BUG-01: `DeletePodcasDeleteOnlyPodcasttEpisodesById`

**Location:** `controllers/podcast.go:151-171`

```go
func DeletePodcastEpisodesById(c *gin.Context) {         // line 151 - correct name
  // ... identical body ...
}
func DeletePodcasDeleteOnlyPodcasttEpisodesById(c *gin.Context) {  // line 162 - typo name
  // ... IDENTICAL body ...
}
```

**Fix:** Remove lines 162-171 (the typo function). Keep lines 151-160.

**Verification:** After removal, `grep -n "DeletePodcasDeleteOnlyPodcastt" controllers/podcast.go` should return no results.

### 3.2 BUG-02: `removeStartingSlash` Logic Reversed

**Location:** `main.go:44-50`

```go
"removeStartingSlash": func(raw string) string {
    fmt.Println(raw)              // BUG-03: debug print
    if string(raw[0]) == "/" {
        return raw                // BUG-02: returns UNCHANGED when slash present
    }
    return "/" + raw              // BUG-02: ADDS slash when not present
},
```

Current behavior:
- Input `"/foo"` → output `"/foo"` (slash NOT removed)
- Input `"foo"` → output `"/foo"` (slash ADDED)

**Expected behavior:**
- Input `"/foo"` → output `"foo"` (slash removed)
- Input `"foo"` → output `"foo"` (unchanged)

**Fix:**
```go
"removeStartingSlash": func(raw string) string {
    if len(raw) > 0 && raw[0] == '/' {
        return raw[1:]
    }
    return raw
},
```

**Verification:**
- `go test -run TestRemoveStartingSlash ./main_test.go` should pass
- Template output for `"/foo/bar"` should be `"foo/bar"`

### 3.3 BUG-03: Debug `fmt.Println` in `removeStartingSlash`

**Location:** `main.go:45` — `fmt.Println(raw)`

**Fix:** Remove the `fmt.Println(raw)` line.

**Verification:** After fix, `grep -n 'fmt.Println(raw)' main.go` returns nothing.

---

## 4. Test File Structure

Per `02-CONTEXT.md` D-07, create minimum test files:

```
db/db_test.go                    # DB CRUD functions
service/podcastService_test.go   # Service layer core functions
main_test.go                     # Template functions (removeStartingSlash)
```

### 4.1 `db/db_test.go` — Core Queries

Test plan:
- `TestCreatePodcast` — insert a podcast, verify UUID generated
- `TestGetPodcastByURL` — lookup by URL after insert
- `TestGetAllPodcasts` — insert multiple, verify count and order
- `TestDeletePodcastById` — insert then delete, verify gone
- `TestGetAllPodcastItemsByPodcastId` — insert podcast + items, query items

### 4.2 `service/podcastService_test.go` — Business Logic

Test plan:
- `TestParseOpml` — XML parsing (no DB needed)
- `TestGetPodcastById` — DB-dependent, via test DB
- `TestDeletePodcastEpisodes` — verify items deleted, podcast stays
- `TestDeletePodcast` — verify cascade delete

### 4.3 `main_test.go` — Template Functions

Test plan:
- `TestRemoveStartingSlash` — verify all cases with leading slash, without, empty string

---

## 5. Technical Debt Noted

### 5.1 `NatualTime` Typo
`service/naturaltime.go:9` — function name `NatualTime` should be `NaturalTime`. Used in `main.go:62`. This is a fourth trivial bug that should be fixed for consistency, though not in ROADMAP requirements.

### 5.2 `DeletePodcastById` in DB Layer
`db/dbfunctions.go:154` — function name collides with controller name. Go allows cross-package name duplication, but it's confusing.

### 5.3 `db.DB` Global
All tests must override or initialize `db.DB`. This is a structural limitation we work around in Phase 2, fix structurally in Phase 4.

---

## 6. Success Criteria Mapping

| Success Criterion | How Verified |
|---|---|
| `go test ./...` runs | Run the command after test files created |
| Podcast delete handler has single correctly-named function | `grep` for typo returns 0 results |
| `removeStartingSlash` removes leading slash | `main_test.go` passes with `t.Errorf` |
| No debug `fmt.Println` in `removeStartingSlash` | `grep` returns 0 results |
| Service and DB tests pass | `go test ./...` exits 0 |

---

## RESEARCH COMPLETE

Path: `/Users/rentamac/Dev/podgrab/.planning/phases/02-test-framework-code-quality/02-RESEARCH.md`

Summary: Phase 2 needs 3 test files (`db/db_test.go`, `service/podcastService_test.go`, `main_test.go`), 3 bug fixes (remove typo handler, fix `removeStartingSlash`, remove debug print), and SQLite `:memory:` test DB setup with global `db.DB` override for per-test isolation.

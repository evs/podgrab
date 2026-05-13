# Phase 3 Research: Correctness & Concurrency Fixes

**Researched:** 2026-05-13
**Status:** Ready for planning

---

## BUG-04: Download Concurrency Limit Ignored

**Current behavior (service/podcastService.go:515-547):**
`DownloadMissingEpisodes()` loops through all episodes to download, launching one goroutine per episode inside a `sync.WaitGroup`. The `index%setting.MaxDownloadConcurrency == 0` check only calls `wg.Wait()` when the index is a multiple of the limit — this means:
- First N episodes launch immediately (unbounded)
- wg.Wait() blocks until ALL N complete
- Then next N launch, etc.
- This is **batching**, not **concurrency limiting**. If there are 1000 episodes, 1000 goroutines launch initially (the first `wg.Wait()` only triggers when index==0, which is the first iteration).

Wait, re-reading line 540: `if index%setting.MaxDownloadConcurrency == 0 { wg.Wait() }`
- index=0: 0%N==0 → wg.Wait() immediately. But wg only has 1 goroutine from this iteration (line 533 added it, but execution could reach 540 before goroutine finishes? No, goroutines start asynchronously, so wg.Wait() at index=0 would block after 1 goroutine, not before the first batch launches. Actually, let's trace:
  - index 0: wg.Add(1), launch goroutine, check 0%N==0 → wg.Wait(). This blocks until goroutine 0 finishes. 
  - index 1: wg.Add(1), launch goroutine, check 1%N==0? false. Continue.
  - ...
  - index N: wg.N has N goroutines active, check N%N==0 → wg.Wait(). Blocks until all N finish.
  - index N+1: etc.

So it IS batching, not true concurrency limiting. The first batch runs sequentially (index=0 blocks immediately). Subsequent indices 1..N-1 accumulate. At index N, all N must complete before continuing. This means:
- **First episode is effectively serial** (waits at index 0)
- **Peak concurrency: N** (episodes 1 through N-1 in flight when we reach index N)
- **But this is still wrong**: if N is large, you still have N goroutines. And if MaxDownloadConcurrency is 1, index=0 waits for goroutine 0, then index=1 launches and at index=1, 1%1==0 so waits immediately, so it's serial.

Actually, I need to read this more carefully. The issue in CONTEXT.md says "No semaphore or worker pool". Current code does have a batching mechanism but it doesn't properly limit concurrency. The bug is that `index%N==0` is a crude batching approach that doesn't control in-flight goroutines properly.

**Root cause:** Using modulo-based batching instead of a bounded worker pool or semaphore.

**Recommended fix:** Replace with a buffered channel semaphore:
```go
sem := make(chan struct{}, setting.MaxDownloadConcurrency)
for _, item := range *data {
    sem <- struct{}{} // acquire
    wg.Add(1)
    go func(item db.PodcastItem) {
        defer wg.Done()
        defer func() { <-sem }() // release
        // download...
    }(item)
}
wg.Wait()
```

This ensures at most `MaxDownloadConcurrency` goroutines are active at any time.

**Alternative:** Use `golang.org/x/sync/semaphore` pkg. But since we're minimizing deps, buffered channel is idiomatic Go.

**Test approach:**
- Mock `Download` and count concurrent invocations
- Verify count never exceeds `MaxDownloadConcurrency`
- Test with `MaxDownloadConcurrency = 1, 3, 5`

---

## BUG-05: RSS Date Parsing Only Tries RFC1123 Variants

**Current behavior (service/podcastService.go:279-301):**
Tries 5 hardcoded formats:
1. `time.RFC1123Z` (Mon, 02 Jan 2006 15:04:05 -0700)
2. `time.RFC1123` (Mon, 02 Jan 2006 15:04:05 MST)
3. Modified single-digit: `Mon, 2 Jan 2006 15:04:05 MST`
4. Modified single-digit with numeric zone: `Mon, 2 Jan 2006 15:04:05 -0700`
5. Modified single-digit with zero-padded day: `Mon, 02 Jan 2006 15:04:05 -0700`

The last two formats are redundant (both try numeric zone). Format 4 and 5 differ only in day padding but format 4 uses `modifiedRFC1123Z` name incorrectly.

**Missing formats seen in the wild:**
- ISO 8601 / RFC3339: `2006-01-02T15:04:05Z` or `2006-01-02T15:04:05-07:00`
- Date-only: `2006-01-02`
- Custom: `Monday, January 2, 2006` (Apple Podcasts sometimes uses this)

**Current code silently ignores parse errors** (blank identifier `_`), resulting in `0001-01-01` dates.

**Root cause:** No fallback to `time.RFC3339` / `time.RFC3339Nano` and no systematic date parsing.

**Recommended fix:** Try `time.Parse(time.RFC3339, ...)` after RFC1123 attempts, then `time.Parse(time.RFC3339Nano, ...)`, then a custom list. Consider `github.com/araddon/dateparse` but minimizing deps is preferred. A helper function with a slice of formats is sufficient:

```go
var rssDateFormats = []string{
    time.RFC1123,
    time.RFC1123Z,
    "Mon, 2 Jan 2006 15:04:05 MST",
    "Mon, 2 Jan 2006 15:04:05 -0700",
    "Mon, 02 Jan 2006 15:04:05 -0700",
    time.RFC3339,
    time.RFC3339Nano,
    "2006-01-02T15:04:05Z",
    "2006-01-02",
}

func parseRSSDate(s string) (time.Time, error) {
    s = strings.TrimSpace(s)
    for _, format := range rssDateFormats {
        if t, err := time.Parse(format, s); err == nil {
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf("cannot parse date: %q", s)
}
```

**Test approach:**
- Table-driven test with all known RSS date formats
- Test that invalid dates return error (or at minimum, don't silently succeed)
- Test that `parseRSSDate` returns non-zero time for real-world RSS strings

---

## BUG-06: Database Initialization Returns Nil DB on Failure

**Current behavior (db/db.go:17-33):**
```go
func Init() (*gorm.DB, error) {
    // ...
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }
    localDB, _ := db.DB()
    // ...
}
```

There are two issues:
1. `db.DB()` error is ignored with `_`
2. Even though `gorm.Open` failure calls `log.Fatalf`, there's no validation that `localDB` is non-nil before calling `SetMaxIdleConns` and `Exec`, which would panic on nil.

However, looking more carefully, `gorm.Open` failure calls `log.Fatalf`, which exits the process. So issue #1 is moot unless `gorm.Open` succeeds but `db.DB()` fails (unlikely with SQLite but possible with connection pool exhaustion or bad PRAGMAs).

Wait — CONTEXT.md says "Database initialization failure returns nil DB silently, causing nil pointer dereferences later". Looking at code: `db.DB()` returns `(*sql.DB, error)`. If `db.DB()` returns nil (error non-nil), then `localDB.SetMaxIdleConns(10)` will panic.

Currently `gorm.Open(sqlite.Open(...))` with `&gorm.Config{}` — if the DB file is in a directory that doesn't exist, `gorm.Open` will create it? Or fail? SQLite auto-creates files. But if CONFIG dir doesn't exist or is not writable, `gorm.Open` may succeed (SQLite creates file) but write operations fail later. However, `localDB` could be nil if `gorm.Open` returns a valid *gorm.DB but the underlying sql.DB is nil (internal GORM state).

**Root cause:** `db.DB()` error discarded; no nil check on `localDB`.

**Recommended fix:**
```go
localDB, err := db.DB()
if err != nil {
    return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
}
```

Also validate DB can execute a simple query:
```go
if err := localDB.Ping(); err != nil {
    return nil, fmt.Errorf("database ping failed: %w", err)
}
```

**Test approach:**
- Test with a read-only directory / invalid path → expect error
- Test with valid path → expect success and working DB
- Consider: `Init` is called from `main.go` and currently uses `log.Fatal` on error. Change to return error and let main decide.

Wait, looking at main.go... I haven't read it, but the plan should account for how main.go handles `db.Init()` return value.

---

## BUG-07: WebSocket Maps Accessed Without Mutex

**Current behavior (controllers/websockets.go:19-114):**
`activePlayers` and `allConnections` are plain `map[*websocket.Conn]string` accessed from:
- `Wshandler` (HTTP handler goroutine) — reads/writes on connection/disconnection
- `HandleWebsocketMessages` (dedicated goroutine) — reads/writes on broadcast loop
- Multiple concurrent WebSocket connections — each `Wshandler` runs in its own goroutine

**Race conditions:**
1. Concurrent write to `activePlayers` or `allConnections` from multiple `Wshandler` goroutines
2. Concurrent read/write between `HandleWebsocketMessages` and `Wshandler` goroutines
3. Map iteration (`for connection := range allConnections`) while another goroutine writes/deletes

Go's map is NOT safe for concurrent use. The race detector will flag this immediately with `go test -race`.

**Root cause:** Missing `sync.RWMutex` (or `sync.Map`) around map operations.

**Recommended fix:** Wrap all map operations with `sync.RWMutex`:
```go
var (
    activePlayersMu sync.RWMutex
    activePlayers = make(map[*websocket.Conn]string)
    allConnectionsMu sync.RWMutex
    allConnections = make(map[*websocket.Conn]string)
)
```

Or use a single mutex for both maps (simpler, less contention concern for this use case).

Refactor into helper functions to centralize locking:
```go
func registerConnection(conn *websocket.Conn, id string) {
    allConnectionsMu.Lock()
    allConnections[conn] = id
    allConnectionsMu.Unlock()
}
```

**Alternative: sync.Map**
`sync.Map` is designed for this pattern but uses `interface{}` for keys/values, which is less type-safe. Given only two maps and simple operations, `sync.RWMutex` is cleaner.

**Test approach:**
- Add `-race` flag to `go test` in the testing workflow
- Write a test that concurrently calls register/unregister on maps and run with `-race`
- A simple test: simulate multiple connections being added/removed from different goroutines

Wait, but testing websockets with the race detector requires actual WebSocket connections or at least the map mutation functions. Isolating the map operations into testable functions helps.

**Note on WebSocket library:** The nhooyr.io/websocket package handles connection safety internally, but the user's map bookkeeping is the problem.

---

## Validation Strategy

### Build & Type Safety
```bash
go build ./...
go vet ./...
```

### Race Detection
```bash
go test -race ./controllers/...
go test -race ./service/...
go test -race ./db/...
```

### Unit Tests
```bash
go test ./...
```

### Specific Verification
1. **BUG-04**: Run `go test -v ./service/... -run TestDownloadConcurrency` — verify concurrency never exceeds limit
2. **BUG-05**: Run `go test -v ./service/... -run TestParseRSSDate` — verify all known formats parse correctly and bad formats return error
3. **BUG-06**: Run `go test -v ./db/... -run TestDBInit` — verify error returned on bad path, success on valid path
4. **BUG-07**: Run `go test -race -v ./controllers/... -run TestWebSocketMaps` — verify no data races detected

---

## Risks

1. **Download semaphore**: Changing the concurrency model could subtly alter download ordering. Document that ordering is not guaranteed.
2. **Date parsing**: Adding new formats means episodes previously getting `0001-01-01` may now parse correctly, changing sort order. This is a bugfix, but user-visible.
3. **WebSocket refactor**: Locking adds latency. For the small scale of this app, negligible. But ensure `RWMutex` is used (reads can be concurrent).
4. **DB init**: Adding `Ping()` could slow down startup by ~1 RTT to the SQLite file (minimal). If `CONFIG` directory doesn't exist, SQLite auto-creates it but `Ping` verifies it works.

## Recommended Libraries

- None needed — all fixes use standard library (`sync`, `time`, `database/sql`).
- If date parsing becomes problematic, `github.com/araddon/dateparse` is a robust fallback (add as new dependency).

---

*Phase: 03-correctness-concurrency-fixes*
*Research gathered: 2026-05-13*

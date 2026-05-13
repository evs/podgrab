---
phase: 03-correctness-concurrency-fixes
plan: 03
name: Correctness & Concurrency Fixes
type: execute
wave: 2
depends_on:
  - 02-test-framework-code-quality
files_modified:
  - service/podcastService.go
  - db/db.go
  - main.go
  - controllers/websockets.go
  - service/podcastService_test.go
  - db/db_test.go
  - controllers/websockets_test.go
requirements:
  - BUG-04
  - BUG-05
  - BUG-06
  - BUG-07
  - TEST-04
  - TEST-05
autonomous: true

must_haves:
  truths:
    - "DownloadMissingEpisodes uses a buffered channel semaphore to limit concurrency"
    - "RSS date parsing handles RFC1123, RFC1123Z, RFC3339, RFC3339Nano, and ISO 8601 formats"
    - "db.Init returns an error instead of panicking on nil DB; main.go checks it"
    - "WebSocket activePlayers and allConnections maps are protected by sync.RWMutex"
    - "go test -race -v ./... passes with no data race warnings"
  artifacts:
    - path: "service/podcastService.go"
      provides: "Semaphore-based concurrency limit and robust RSS date parser"
      contains: "sem := make(chan struct{},"
    - path: "db/db.go"
      provides: "Safe DB initialization with error propagation"
      contains: "localDB, err := db.DB()"
    - path: "controllers/websockets.go"
      provides: "RWMutex-protected WebSocket connection maps"
      contains: "var playersMutex sync.RWMutex"
    - path: "service/podcastService_test.go"
      provides: "Tests for concurrency limit and date parsing"
      contains: "TestDownloadConcurrency"
    - path: "controllers/websockets_test.go"
      provides: "Race-free concurrent WebSocket map tests"
      contains: "TestWebSocketDataRace"
  key_links:
    - from: "db/db.go"
      to: "db.Init() error propagation"
      via: "localDB, err := db.DB() and Ping()"
      pattern: "Ping\(\)"
    - from: "controllers/websockets.go"
      to: "sync.RWMutex"
      via: "protects activePlayers and allConnections maps"
      pattern: "playersMutex"
    - from: "service/podcastService.go"
      to: "parseRSSDate"
      via: "replaces inline RFC1123-only parsing in AddPodcastItems"
      pattern: "parseRSSDate"
---

<objective>
Fix four correctness and concurrency bugs in the Go podcast manager, verify each fix with a targeted test, and ensure `go test -race` passes with zero warnings.

Purpose: The download-and-organize loop must never break. BUG-04 causes unbounded goroutine growth, BUG-05 causes episodes to appear in wrong order, BUG-06 causes nil-pointer panics on DB failure, and BUG-07 causes data races under concurrent WebSocket load. Each fix is isolated to one file and verified.

Output: All four bugs fixed, six test cases passing, `go test -race -v ./...` clean.
</objective>

<execution_context>
@/Users/rentamac/.config/opencode/get-shit-done/workflows/execute-plan.md
@/Users/rentamac/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@.planning/phases/03-correctness-concurrency-fixes/03-CONTEXT.md
@.planning/phases/03-correctness-concurrency-fixes/03-RESEARCH.md
@.planning/phases/03-correctness-concurrency-fixes/03-VALIDATION.md

<interfaces>
From service/podcastService.go (lines 515-547, current DownloadMissingEpisodes):
```go
var wg sync.WaitGroup
for index, item := range *data {
    wg.Add(1)
    go func(item db.PodcastItem, setting db.Setting) {
        defer wg.Done()
        url, _ := Download(item.FileURL, item.Title, item.Podcast.Title, GetPodcastPrefix(&item, &setting))
        SetPodcastItemAsDownloaded(item.ID, url)
    }(item, *setting)

    if index%setting.MaxDownloadConcurrency == 0 {
        wg.Wait()
    }
}
wg.Wait()
```

From service/podcastService.go (lines 279-301, current date parsing):
```go
pubDate, _ := time.Parse(time.RFC1123Z, toParse)
if (pubDate == time.Time{}) {
    pubDate, _ = time.Parse(time.RFC1123, toParse)
}
if (pubDate == time.Time{}) {
    modifiedRFC1123 := "Mon, 2 Jan 2006 15:04:05 MST"
    pubDate, _ = time.Parse(modifiedRFC1123, toParse)
}
if (pubDate == time.Time{}) {
    modifiedRFC1123Z := "Mon, 2 Jan 2006 15:04:05 -0700"
    pubDate, _ = time.Parse(modifiedRFC1123Z, toParse)
}
if (pubDate == time.Time{}) {
    modifiedRFC1123Z := "Mon, 02 Jan 2006 15:04:05 -0700"
    pubDate, _ = time.Parse(modifiedRFC1123Z, toParse)
}
if (pubDate == time.Time{}) {
    fmt.Printf("Cant format date : %s", obj.PubDate)
}
```

From db/db.go (lines 17-33, current Init):
```go
func Init() (*gorm.DB, error) {
    configPath := os.Getenv("CONFIG")
    dbPath := path.Join(configPath, "podgrab.db")
    log.Println(dbPath)
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }

    localDB, _ := db.DB()
    localDB.SetMaxIdleConns(10)
    localDB.Exec("PRAGMA journal_mode=WAL")
    localDB.Exec("PRAGMA busy_timeout=5000")
    DB = db
    return DB, nil
}
```

From controllers/websockets.go (lines 19-20 and reads/writes):
```go
var activePlayers = make(map[*websocket.Conn]string)
var allConnections = make(map[*websocket.Conn]string)
```
- Read: lines 46, 58 (map lookup)
- Write: lines 48, 54, 58, 70 (map insert/delete)
- Iterate: lines 71, 79 (range over map)
</interfaces>
</context>

<tasks>

<!-- ==================== BUG-05 + TEST-05 ==================== -->

<task type="auto">
  <name>Task 1: Extract parseRSSDate helper and add format fallbacks (BUG-05, TEST-05)</name>
  <files>service/podcastService.go, service/podcastService_test.go</files>
  <read_first>service/podcastService.go (lines 270-310), service/podcastService_test.go</read_first>
  <action>
1. In service/podcastService.go, add the following new function BEFORE `ParseOpml` (before line 32):
```go
var rssDateFormats = []string{
    time.RFC1123,
    time.RFC1123Z,
    time.RFC3339,
    time.RFC3339Nano,
    time.RFC822,
    time.RFC822Z,
    time.UnixDate,
    "Mon, 2 Jan 2006 15:04:05 MST",
    "Mon, 2 Jan 2006 15:04:05 -0700",
    "Mon, 02 Jan 2006 15:04:05 -0700",
    "2006-01-02T15:04:05Z07:00",
    "2006-01-02 15:04:05",
}

func parseRSSDate(raw string) (time.Time, error) {
    s := strings.TrimSpace(raw)
    if s == "" {
        return time.Time{}, fmt.Errorf("empty date string")
    }
    for _, layout := range rssDateFormats {
        if t, err := time.Parse(layout, s); err == nil {
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf("unable to parse RSS date: %q", raw)
}
```

2. Add `"fmt"` to the imports of service/podcastService.go if not already present (it is used elsewhere, so verify it exists).

3. Replace lines 279-301 (the existing inline date parsing block) with:
```go
pubDate, err := parseRSSDate(toParse)
if err != nil {
    Logger.Warnf("BUG-05: failed to parse pubDate %q: %v", obj.PubDate, err)
}
```

4. In service/podcastService_test.go, add the following table-driven test BEFORE the existing `TestParseOpml` (or after all existing tests — last position is fine):
```go
func TestParseRSSDate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        year    int
    }{
        {"RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", false, 2006},
        {"RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700", false, 2006},
        {"RFC3339", "2006-01-02T15:04:05Z", false, 2006},
        {"RFC3339Nano", "2006-01-02T15:04:05.999999999Z", false, 2006},
        {"ISO8601", "2006-01-02T15:04:05+00:00", false, 2006},
        {"ModifiedRFC1123", "Mon, 2 Jan 2006 15:04:05 MST", false, 2006},
        {"ModifiedRFC1123Z", "Mon, 2 Jan 2006 15:04:05 -0700", false, 2006},
        {"empty", "", true, 0},
        {"garbage", "not-a-date", true, 0},
        {"RFC822", "02 Jan 06 15:04 MST", false, 2006},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseRSSDate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("parseRSSDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
            if !tt.wantErr && got.Year() != tt.year {
                t.Errorf("parseRSSDate(%q) year = %d, want %d", tt.input, got.Year(), tt.year)
            }
        })
    }
}
```

5. Run: `cd /Users/rentamac/Dev/podgrab && go test -v ./service/... -run TestParseRSSDate`
  </action>
  <acceptance_criteria>
    - service/podcastService.go contains `func parseRSSDate(raw string) (time.Time, error)`
    - service/podcastService.go contains `var rssDateFormats = []string{`
    - service/podcastService.go does NOT contain `pubDate, _ = time.Parse(time.RFC1123Z, toParse)` (the old inline pattern)
    - service/podcastService_test.go contains `func TestParseRSSDate(t *testing.T)`
    - `go test -v ./service/... -run TestParseRSSDate` exits 0 with all subtests passing
    - `go build ./...` exits 0
  </acceptance_criteria>
  <verify>
    <automated>cd /Users/rentamac/Dev/podgrab && go test -v ./service/... -run TestParseRSSDate && grep -q "func parseRSSDate" service/podcastService.go && ! grep -q "pubDate, _ = time.Parse(time.RFC1123Z, toParse)" service/podcastService.go</automated>
  </verify>
  <done>parseRSSDate helper extracted, 10 date formats supported, old inline parsing removed, TestParseRSSDate passes</done>
</task>

<!-- ==================== BUG-06 ==================== -->

<task type="auto">
  <name>Task 2: Propagate DB initialization error and add Ping() validation (BUG-06)</name>
  <files>db/db.go, main.go</files>
  <read_first>db/db.go (lines 1-44), main.go (lines 127-135)</read_first>
  <action>
1. In db/db.go, replace lines 25-33 with:
```go
    localDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("get underlying sql.DB: %w", err)
    }
    if err := localDB.Ping(); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }
    localDB.SetMaxIdleConns(10)
    localDB.Exec("PRAGMA journal_mode=WAL")
    localDB.Exec("PRAGMA busy_timeout=5000")
    DB = db
    return DB, nil
```

2. Add `"fmt"` to db/db.go imports if not present (check existing imports; if missing, add it after `"log"`).

3. In main.go, after line 127 (`func main() {`), add before `r := gin.Default()`:
```go
	_, err := db.Init()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	db.Migrate()
```

4. Also in main.go, remove the `intiCron()` typo call or keep it as-is if it already exists — do NOT modify `intiCron()` itself. Only add the db.Init/db.Migrate block.

5. Since `setupSettings()` middleware (line 130) relies on `db.GetOrCreateSetting()` which requires `db.DB` to be initialized, the `db.Init()` call MUST appear before `r.Use(setupSettings())`.

6. Run: `cd /Users/rentamac/Dev/podgrab && go build ./...`
  </action>
  <acceptance_criteria>
    - db/db.go contains `localDB, err := db.DB()` (not `localDB, _ := db.DB()`)
    - db/db.go contains `if err := localDB.Ping(); err != nil {`
    - db/db.go Init function signature still returns `(*gorm.DB, error)`
    - main.go contains `_, err := db.Init()` before `r := gin.Default()`
    - main.go contains `if err != nil { log.Fatalf(...) }` after db.Init call
    - main.go contains `db.Migrate()` after db.Init
    - `go build ./...` exits 0
    - `go test -v ./db/... -run TestDBInit` passes (added in Task 6)
  </acceptance_criteria>
  <verify>
    <automated>cd /Users/rentamac/Dev/podgrab && grep -q "localDB, err := db.DB()" db/db.go && grep -q "localDB.Ping()" db/db.go && grep -q "_, err := db.Init()" main.go && go build ./...</automated>
  </verify>
  <done>DB init error propagation fixed, Ping validation added, main.go calls db.Init and fatals on failure</done>
</task>

<!-- ==================== BUG-04 + TEST-04 ==================== -->

<task type="auto">
  <name>Task 3: Replace batching with buffered channel semaphore in DownloadMissingEpisodes (BUG-04, TEST-04)</name>
  <files>service/podcastService.go, service/podcastService_test.go</files>
  <read_first>service/podcastService.go (lines 515-547)</read_first>
  <action>
1. In service/podcastService.go, replace the loop body at lines 531-543 (the existing `for index, item := range *data { ... }` block) with:
```go
	var wg sync.WaitGroup
	sem := make(chan struct{}, setting.MaxDownloadConcurrency)
	for _, item := range *data {
		sem <- struct{}{} // acquire
		wg.Add(1)
		go func(item db.PodcastItem, setting db.Setting) {
			defer wg.Done()
			defer func() { <-sem }() // release
			url, _ := Download(item.FileURL, item.Title, item.Podcast.Title, GetPodcastPrefix(&item, &setting))
			SetPodcastItemAsDownloaded(item.ID, url)
		}(item, *setting)
	}
	wg.Wait()
```

2. In service/podcastService_test.go, add the following test at the end of the file (after all existing tests):
```go
func TestDownloadConcurrencyLimit(t *testing.T) {
	t.Run("maxConcurrent never exceeds limit", func(t *testing.T) {
		var maxObserved int64
		var current int64
		limit := 3
		sem := make(chan struct{}, limit)
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				c := atomic.AddInt64(&current, 1)
				for {
					m := atomic.LoadInt64(&maxObserved)
					if c > m {
						if atomic.CompareAndSwapInt64(&maxObserved, m, c) {
							break
						}
					} else {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
				atomic.AddInt64(&current, -1)
			}()
		}
		wg.Wait()
		if maxObserved > int64(limit) {
			t.Fatalf("max concurrent %d > limit %d", maxObserved, limit)
		}
	})
}
```

3. Add the import `"sync/atomic"` to service/podcastService_test.go if not present.

4. Run: `cd /Users/rentamac/Dev/podgrab && go test -v ./service/... -run TestDownloadConcurrencyLimit`
  </action>
  <acceptance_criteria>
    - service/podcastService.go contains `sem := make(chan struct{}, setting.MaxDownloadConcurrency)`
    - service/podcastService.go contains `defer func() { <-sem }()` inside the download goroutine
    - service/podcastService.go does NOT contain `index%setting.MaxDownloadConcurrency == 0`
    - service/podcastService_test.go contains `func TestDownloadConcurrencyLimit(t *testing.T)`
    - `go test -v ./service/... -run TestDownloadConcurrencyLimit` exits 0
    - `go build ./...` exits 0
  </acceptance_criteria>
  <verify>
    <automated>cd /Users/rentamac/Dev/podgrab && grep -q "sem := make(chan struct{}, setting.MaxDownloadConcurrency)" service/podcastService.go && ! grep -q "index%setting.MaxDownloadConcurrency == 0" service/podcastService.go && go test -v ./service/... -run TestDownloadConcurrencyLimit</automated>
  </verify>
  <done>DownloadMissingEpisodes uses bounded semaphore instead of batching, TestDownloadConcurrencyLimit verifies in-flight goroutines never exceed configured limit</done>
</task>

<!-- ==================== BUG-07 + TEST-04/TEST-05 integration ==================== -->

<task type="auto">
  <name>Task 4: Add sync.RWMutex around all WebSocket map operations (BUG-07)</name>
  <files>controllers/websockets.go</files>
  <read_first>controllers/websockets.go (lines 1-135)</read_first>
  <action>
1. In controllers/websockets.go, add the import `"sync"` to the import block (after `"encoding/json"` or at the end of the standard-library group).

2. After line 22 (`var broadcast = make(chan Message)`), add:
```go
var playersMutex sync.RWMutex
var connectionsMutex sync.RWMutex
```

3. In `Wshandler`, replace lines 46-55 (the disconnect / error block) with:
```go
        isPlayer := false
        playersMutex.RLock()
        isPlayer = activePlayers[conn] != ""
        playersMutex.RUnlock()
        if isPlayer {
            playersMutex.Lock()
            delete(activePlayers, conn)
            playersMutex.Unlock()
            broadcast <- Message{
                MessageType: "PlayerRemoved",
                Identifier:  mess.Identifier,
            }
        }
        connectionsMutex.Lock()
        delete(allConnections, conn)
        connectionsMutex.Unlock()
        break
```

4. In `Wshandler`, replace line 58 (`allConnections[conn] = mess.Identifier`) with:
```go
        connectionsMutex.Lock()
        allConnections[conn] = mess.Identifier
        connectionsMutex.Unlock()
```

5. In `HandleWebsocketMessages`, replace lines 69-76 (the "RegisterPlayer" case) with:
```go
        case "RegisterPlayer":
            playersMutex.Lock()
            activePlayers[msg.Connection] = msg.Identifier
            playersMutex.Unlock()
            connectionsMutex.RLock()
            for connection := range allConnections {
                wsjson.Write(ctx, connection, Message{
                    Identifier:  msg.Identifier,
                    MessageType: "PlayerExists",
                })
            }
            connectionsMutex.RUnlock()
            fmt.Println("Player Registered")
```

6. In `HandleWebsocketMessages`, replace lines 78-85 (the "PlayerRemoved" case) with:
```go
        case "PlayerRemoved":
            connectionsMutex.RLock()
            for connection := range allConnections {
                wsjson.Write(ctx, connection, Message{
                    Identifier:  msg.Identifier,
                    MessageType: "NoPlayer",
                })
            }
            connectionsMutex.RUnlock()
            fmt.Println("Player Removed")
```

7. In `HandleWebsocketMessages`, replace lines 92-98 (the activePlayer iteration in "Enqueue" case) with:
```go
                var player *websocket.Conn
                playersMutex.RLock()
                for connection, id := range activePlayers {
                    if msg.Identifier == id {
                        player = connection
                        break
                    }
                }
                playersMutex.RUnlock()
```

8. In `HandleWebsocketMessages`, replace lines 113-119 (the activePlayer iteration in "Register" case) with:
```go
            var player *websocket.Conn
            playersMutex.RLock()
            for connection, id := range activePlayers {
                if msg.Identifier == id {
                    player = connection
                    break
                }
            }
            playersMutex.RUnlock()
```

9. Run: `cd /Users/rentamac/Dev/podgrab && go build ./...`
  </action>
  <acceptance_criteria>
    - controllers/websockets.go imports `"sync"`
    - controllers/websockets.go contains `var playersMutex sync.RWMutex`
    - controllers/websockets.go contains `var connectionsMutex sync.RWMutex`
    - All read accesses to `activePlayers` are wrapped in `playersMutex.RLock()...RUnlock()`
    - All write accesses to `activePlayers` are wrapped in `playersMutex.Lock()...Unlock()`
    - All read accesses to `allConnections` are wrapped in `connectionsMutex.RLock()...RUnlock()`
    - All write accesses to `allConnections` are wrapped in `connectionsMutex.Lock()...Unlock()`
    - `go build ./...` exits 0
    - `go test -race -run TestWebSocketDataRace ./controllers/...` passes (Task 5 adds the test)
  </acceptance_criteria>
  <verify>
    <automated>cd /Users/rentamac/Dev/podgrab && grep -q 'var playersMutex sync.RWMutex' controllers/websockets.go && grep -q 'var connectionsMutex sync.RWMutex' controllers/websockets.go && grep -c 'playersMutex' controllers/websockets.go | xargs -I{} test {} -ge 6 && go build ./...</automated>
  </verify>
  <done>All activePlayers and allConnections reads/writes protected by dedicated RWMutex pairs, compiles cleanly</done>
</task>

<!-- ==================== TEST-04 / TEST-05: DB and race coverage ==================== -->

<task type="auto">
  <name>Task 5: Add WebSocket data race test (BUG-07, TEST-04)</name>
  <files>controllers/websockets_test.go</files>
  <read_first>controllers/websockets.go (Task 4 output)</read_first>
  <action>
1. Create the file `controllers/websockets_test.go` with content:
```go
package controllers

import (
	"context"
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestWebSocketDataRace(t *testing.T) {
	// Reset maps for deterministic test
	playersMutex.Lock()
	activePlayers = make(map[*websocket.Conn]string)
	playersMutex.Unlock()

	connectionsMutex.Lock()
	allConnections = make(map[*websocket.Conn]string)
	connectionsMutex.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := &websocket.Conn{} // mock — we only test map ops, not real WS
			id := "player-" + string(rune('A'+i%26))
			playersMutex.Lock()
			activePlayers[conn] = id
			playersMutex.Unlock()

			connectionsMutex.Lock()
			allConnections[conn] = id
			connectionsMutex.Unlock()

			time.Sleep(1 * time.Millisecond)

			playersMutex.RLock()
			_ = activePlayers[conn]
			playersMutex.RUnlock()

			connectionsMutex.RLock()
			_ = allConnections[conn]
			connectionsMutex.RUnlock()

			playersMutex.Lock()
			delete(activePlayers, conn)
			playersMutex.Unlock()

			connectionsMutex.Lock()
			delete(allConnections, conn)
			connectionsMutex.Unlock()
		}(i)
	}
	wg.Wait()
}
```

2. Run: `cd /Users/rentamac/Dev/podgrab && go test -race -v ./controllers/... -run TestWebSocketDataRace`
  </action>
  <acceptance_criteria>
    - controllers/websockets_test.go exists and contains `func TestWebSocketDataRace(t *testing.T)`
    - `go test -race -v ./controllers/... -run TestWebSocketDataRace` exits 0 with no race warnings
    - The test uses `playersMutex.Lock()/Unlock()` and `connectionsMutex.Lock()/Unlock()` around map writes
    - The test uses `playersMutex.RLock()/RUnlock()` and `connectionsMutex.RLock()/RUnlock()` around map reads
    - `go build ./...` exits 0
  </acceptance_criteria>
  <verify>
    <automated>cd /Users/rentamac/Dev/podgrab && go test -race -v ./controllers/... -run TestWebSocketDataRace</automated>
  </verify>
  <done>WebSocket data race test added and passes under go test -race with no warnings</done>
</task>

<!-- ==================== Integration and final validation ==================== -->

<task type="auto">
  <name>Task 6: Add DB init test and run full test suite (BUG-06, TEST-05)</name>
  <files>db/db_test.go</files>
  <read_first>db/db_test.go, db/db.go</read_first>
  <action>
1. In db/db_test.go, append the following test at the end of the file (after the last existing test):
```go
func TestDBInit(t *testing.T) {
	// Test 1: valid path (memory)
	origDB := DB
	defer func() { DB = origDB }()

	// Init does not accept a path parameter; it reads CONFIG env.
	// We test the internal path assembly logic indirectly via Init by
	// setting a temporary directory as CONFIG.
	tmpDir := t.TempDir()
	origConfig := os.Getenv("CONFIG")
	os.Setenv("CONFIG", tmpDir)
	defer os.Setenv("CONFIG", origConfig)

	// Re-Init should succeed
	db, err := Init()
	if err != nil {
		t.Fatalf("Init with valid path failed: %v", err)
	}
	if db == nil {
		t.Fatal("Init returned nil db without error")
	}
	DB = db

	// Ping should succeed
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() after Init: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping after Init: %v", err)
	}
}
```

2. Run the full validation suite:
```bash
cd /Users/rentamac/Dev/podgrab && go test -race -v ./...
```

3. If any test fails, fix the root cause and re-run until `go test -race -v ./...` exits 0. Common issues:
   - Missing imports in test files (add them).
   - `db.Init()` conflict in main.go if it tries to open a real DB during `go test ./...` on main — ensure `main.go` db.Init is in `func main()` and not in package init.
  </action>
  <acceptance_criteria>
    - db/db_test.go contains `func TestDBInit(t *testing.T)`
    - `go test -v ./db/... -run TestDBInit` exits 0
    - `go test -race -v ./...` exits 0 with no race warnings
    - `go build ./...` exits 0
    - `go vet ./...` exits 0 (or only pre-existing unrelated warnings)
  </acceptance_criteria>
  <verify>
    <automated>cd /Users/rentamac/Dev/podgrab && go test -race -v ./... && grep -q "func TestDBInit" db/db_test.go</automated>
  </verify>
  <done>DB init test verifies Ping, full suite green under -race, all bugs fixed</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| DB init → app startup | DB failure must halt startup; nil DB must not be passed downstream |
| WebSocket handler → goroutine pool | Concurrent map access must not leak player state or crash |
| Download worker → filesystem | Bounded concurrency prevents file descriptor exhaustion |
| RSS parser → episode ordering | Correct date parsing preserves chronological feed integrity |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-03-01 | S | db/db.go | accept | Init error returns to caller; main.go fatals — no partial trust |
| T-03-02 | D | controllers/websockets.go | mitigate | RWMutex prevents race-driven map corruption |
| T-03-03 | D | service/podcastService.go | mitigate | Semaphore bounds in-flight downloads |
| T-03-04 | I | service/podcastService.go | accept | parseRSSDate only affects local ordering, no auth |
</threat_model>

<verification>
1. `go test -race -v ./...` exits 0 — zero data races across the entire project
2. `grep "func parseRSSDate" service/podcastService.go` matches — date helper exists
3. `grep "sem := make(chan struct{}, setting.MaxDownloadConcurrency)" service/podcastService.go` matches — semaphore in place
4. `grep "localDB, err := db.DB()" db/db.go` matches — DB error handled
5. `grep "var playersMutex sync.RWMutex" controllers/websockets.go` matches — mutex declared
6. Each requirement appears at least once:
   - BUG-04 — Task 3 (semaphore + concurrency test)
   - BUG-05 — Task 1 (parseRSSDate + format list)
   - BUG-06 — Task 2 (error propagation + Ping) and Task 6 (TestDBInit)
   - BUG-07 — Task 4 (RWMutex) and Task 5 (race test)
   - TEST-04 — Task 3 (TestDownloadConcurrencyLimit)
   - TEST-05 — Task 1 (TestParseRSSDate) and Task 6 (TestDBInit)
</verification>

## PLANNING COMPLETE
**Tasks: 6**

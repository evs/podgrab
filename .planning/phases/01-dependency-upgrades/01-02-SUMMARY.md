---
phase: 01-dependency-upgrades
plan: 02
subsystem: infra
tags: [go, golang, dependencies, jwt-go, uuid, cron, websocket, replace]

# Dependency graph
requires:
  - phase: 01
    plan: 01
    provides: Go 1.24+ toolchain with ioutil removal
provides:
  - Google UUID v1.6.0 for model primary keys
  - robfig/cron/v3 for scheduled background tasks
  - coder/websocket (nhooyr.io/websocket) for real-time player sync
  - Removal of 4 abandoned/unused dependencies
affects: [dependency-upgrades, bug-fixes, error-handling]

# Tech tracking
tech-stack:
  added: [github.com/google/uuid v1.6.0, github.com/robfig/cron/v3 v3.0.1, nhooyr.io/websocket v1.8.17]
  patterns: [cron expressions for scheduled tasks, wsjson.Read/Write for WebSocket messaging, UUID generation via google/uuid]

key-files:
  created: []
  modified:
    - go.mod
    - go.sum
    - db/base.go
    - main.go
    - controllers/websockets.go

key-decisions:
  - "Removed dgrijalva/jwt-go entirely — confirmed unused, no source references existed"
  - "Swapped satori/go.uuid → google/uuid — uuid.NewV4().String() → uuid.New().String() (both produce v4 UUIDs)"
  - "Swapped jasonlvhit/gocron → robfig/cron/v3 — converted minute-based intervals to 5-field cron expressions"
  - "Swapped gorilla/websocket → coder/websocket — replaced Upgrader with Accept(), ReadJSON/WriteJSON with wsjson.Read/Write"
  - "Wrapped service functions returning errors in func() closures for cron.AddFunc compatibility"
  - "Non-blocking cron.Start() replaces blocking gocron.Start() channel — Gin's r.Run() blocks main goroutine"
  - "Set InsecureSkipVerify: true on WebSocket AcceptOptions to match current gorilla behavior (no origin check)"

patterns-established:
  - "Cron scheduling uses standard 5-field expressions (*/N * * * *) instead of fluent gocron API"
  - "WebSocket upgrade uses websocket.Accept() with AcceptOptions instead of Upgrader.Upgrade()"
  - "WebSocket message read/write uses wsjson.Read/Write with context.Background() instead of ReadJSON/WriteJSON"

requirements-completed: [DEPS-03, DEPS-04, DEPS-05, DEPS-07]

# Metrics
duration: 8min
completed: 2026-05-12
---

# Phase 1 Plan 02: Remove Unused jwt-go + Swap Abandoned Libraries Summary

**Four abandoned dependencies removed — jwt-go dropped, UUID/cron/websocket swapped with maintained alternatives, project compiles clean**

## Performance

- **Duration:** 8 min
- **Started:** 2026-05-12T09:11:48Z
- **Completed:** 2026-05-12T09:19:50Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments

- Removed unused dgrijalva/jwt-go dependency (confirmed no source references)
- Replaced satori/go.uuid v1.2.0 with google/uuid v1.6.0 (UUID generation preserved)
- Replaced jasonlvhit/gocron v0.0.1 with robfig/cron/v3 v3.0.1 (all scheduled tasks converted to cron expressions)
- Replaced gorilla/websocket v1.4.2 with nhooyr.io/websocket v1.8.17 (WebSocket upgrade and messaging adapted)
- Fixed PlayerRemoved log message (was incorrectly labeled "Player Registered")
- All four abandoned/unmaintained dependencies eliminated from go.mod and source

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove dgrijalva/jwt-go and swap satori/go.uuid → google/uuid** - `5abb4db` (feat)
2. **Task 2: Swap jasonlvhit/gocron → robfig/cron/v3** - `ddd64eb` (feat)
3. **Task 3: Swap gorilla/websocket → coder/websocket** - `c669825` (feat)

## Files Created/Modified

- `go.mod` - Removed jwt-go, satori/go.uuid, gocron, gorilla/websocket; added google/uuid, robfig/cron/v3, nhooyr.io/websocket
- `go.sum` - Updated module graph after each dependency swap
- `db/base.go` - Changed import from satori/go.uuid to google/uuid; uuid.NewV4().String() → uuid.New().String()
- `main.go` - Replaced gocron scheduling with robfig/cron/v3 cron expressions; non-blocking c.Start()
- `controllers/websockets.go` - Replaced gorilla/websocket with nhooyr.io/websocket; Accept(), wsjson.Read/Write, context.Background()

## Decisions Made

- Used `uuid.New()` (not `uuid.NewV4()`) — google/uuid's `New()` returns v4 UUIDs by default, matching the original behavior
- Wrapped service functions that return errors in `func()` closures for `cron.AddFunc` compatibility — errors are silently discarded (matching gocron's behavior)
- Used `context.Background()` for wsjson operations — Phase 3 will add proper context handling with request-scoped timeouts
- Kept `InsecureSkipVerify: true` on WebSocket AcceptOptions to match current gorilla behavior (no origin check) — security improvement deferred
- Non-blocking `c.Start()` replaces blocking `<-gocron.Start()` — Gin's `r.Run()` already blocks the main goroutine
- Fixed incorrect "Player Registered" log message in PlayerRemoved handler → "Player Removed"

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Cron function signature mismatch**
- **Found during:** Task 2 implementation
- **Issue:** service functions like `RefreshEpisodes() error` don't match `cron.AddFunc`'s `func()` signature
- **Fix:** Wrapped each service call in `func() { service.Func() }` closures to discard error returns
- **Files modified:** main.go
- **Commit:** ddd64eb

**2. [Rule 1 - Bug] Fixed incorrect PlayerRemoved log message**
- **Found during:** Task 3 rewrite
- **Issue:** Original gorilla implementation logged "Player Registered" in the PlayerRemoved case
- **Fix:** Changed to "Player Removed" in the rewritten code
- **Files modified:** controllers/websockets.go
- **Commit:** c669825

## Known Stubs

None — all functionality wired to real implementations.

## Threat Flags

None — all threat model dispositions were accepted as planned.

---
*Phase: 01-dependency-upgrades*
*Completed: 2026-05-12*
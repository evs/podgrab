---
phase: 01-dependency-upgrades
plan: 01
subsystem: infra
tags: [go, golang, ioutil, go-mod, dependency-upgrade]

# Dependency graph
requires:
  - phase: none
    provides: N/A (first plan)
provides:
  - Go 1.24+ toolchain compatibility
  - Removal of io/ioutil deprecation warnings
  - Clean build on Go 1.24+
affects: [dependency-upgrades, bug-fixes, error-handling]

# Tech tracking
tech-stack:
  added: []
  patterns: [io.ReadAll replaces ioutil.ReadAll, os.WriteFile replaces ioutil.WriteFile]

key-files:
  created: []
  modified:
    - go.mod
    - go.sum
    - service/podcastService.go
    - service/fileService.go

key-decisions:
  - "Upgraded to Go 1.24 (not 1.22 minimum) to use the latest stable toolchain"
  - "Replaced ioutil.ReadAll with io.ReadAll (identical behavior, different import path)"
  - "Replaced ioutil.WriteFile with os.WriteFile (identical behavior, different import path)"

patterns-established:
  - "Use io.ReadAll instead of ioutil.ReadAll for HTTP response body reading"
  - "Use os.WriteFile instead of ioutil.WriteFile for file writing"

requirements-completed: [DEPS-01, DEPS-08]

# Metrics
duration: 4min
completed: 2026-05-12
---

# Phase 1 Plan 01: Go Version Upgrade + ioutil Replacement Summary

**Go 1.24 toolchain upgrade with full io/ioutil deprecation removal — project compiles and vets clean on modern Go**

## Performance

- **Duration:** 4 min
- **Started:** 2026-05-12T08:59:11Z
- **Completed:** 2026-05-12T09:03:29Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Go version directive upgraded from 1.15 to 1.24 in go.mod
- All deprecated io/ioutil usage replaced with io.ReadAll and os.WriteFile
- Project builds and passes vet checks on Go 1.24+ (pre-existing vet warnings are unrelated)

## Task Commits

Each task was committed atomically:

1. **Task 1: Upgrade Go version to 1.24+ and fix go.mod** - `b6f1d30` (feat)
2. **Task 2: Replace all io/ioutil usage with modern equivalents** - `3d6ff08` (feat)

## Files Created/Modified
- `go.mod` - Go version directive upgraded from 1.15 to 1.24
- `go.sum` - Updated module graph after go mod tidy
- `service/podcastService.go` - Replaced ioutil.ReadAll with io.ReadAll, swapped import from io/ioutil to io
- `service/fileService.go` - Replaced ioutil.WriteFile with os.WriteFile, removed io/ioutil import

## Decisions Made
- Chose Go 1.24 as target (latest stable, well beyond minimum 1.22 required for ioutil removal)
- Kept pre-existing go vet warnings (non-constant format string, struct tag syntax, Printf formatting) as out of scope — they predate this plan

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- go vet reports 4 pre-existing warnings (model/errors.go format string, model/queryModels.go struct tag syntax, controllers/websockets.go Printf formatting). These are not introduced by this plan and are out of scope per deviation Rule scope boundary.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Go 1.24+ toolchain established, ready for subsequent dependency upgrades (Plan 02)
- Pre-existing vet warnings documented for future cleanup phases

---
*Phase: 01-dependency-upgrades*
*Completed: 2026-05-12*
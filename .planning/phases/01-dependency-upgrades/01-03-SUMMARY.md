---
phase: 01-dependency-upgrades
plan: 03
subsystem: infra
tags: [go, golang, dependencies, gin, gorm, migration, sqlite]

# Dependency graph
requires:
  - phase: 01
    plan: 02
    provides: Abandoned libs removed (uuid, cron, websocket swapped)
provides:
  - Gin v1.12+ HTTP framework
  - GORM v1.26+ with v2 behavioral changes audited and fixed
  - SQLite WAL mode for concurrent read safety
  - All dependencies at modern versions
affects: [dependency-upgrades, bug-fixes, error-handling]

# Tech tracking
tech-stack:
  added: []
  patterns: [gorm.DeletedAt for soft delete, gorm:"primaryKey" camelCase tags, PRAGMA journal_mode=WAL, PRAGMA busy_timeout]

key-files:
  created: []
  modified:
    - go.mod
    - go.sum
    - db/base.go
    - db/db.go

key-decisions:
  - "Upgraded GORM from v1.20.2 to v1.26.0 — v2 behavioral changes audited and fixed"
  - "Changed Base.DeletedAt from *time.Time to gorm.DeletedAt — required for v2 soft delete"
  - "Changed Base.ID tag from sql:\"type:uuid;primary_key\" to gorm:\"type:uuid;primaryKey\" — v2 camelCase convention"
  - "Added SQLite WAL mode and busy_timeout=5000 for concurrent read performance"
  - "Replaced fmt.Println with log.Fatalf on DB init failure — fatal exit on misconfiguration"

patterns-established:
  - "GORM v2 soft delete uses gorm.DeletedAt type — `*time.Time` won't scope queries correctly"
  - "GORM v2 struct tags use camelCase (primaryKey, not primary_key)"
  - "SQLite WAL mode enabled at init for better concurrent read performance"
  - "DB connection failure is now fatal (log.Fatalf) instead of silently returning error"

requirements-completed: [DEPS-02, DEPS-06, DEPS-07]

# Metrics
duration: 10min
completed: 2026-05-12
---

# Phase 1 Plan 03: Upgrade Gin + Remaining Dependencies + GORM Migration Summary

**Gin v1.12, GORM v1.26 with full v2 behavioral audit, all deps modernized, SQLite WAL enabled**

## Performance

- **Duration:** 10 min
- **Started:** 2026-05-12T09:32:33Z
- **Completed:** 2026-05-12T09:43:04Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Upgraded Gin from v1.7.2 to v1.12.0 (major version jump with breaking changes handled)
- Upgraded all remaining non-GORM dependencies to current versions (zap, godotenv, x/crypto, x/net, bluemonday, xmlquery, stringy, html-strip-tags-go, gin-contrib/location, multierr)
- Upgraded GORM from v1.20.2 to v1.26.0 with full v2 migration audit
- Fixed Base.DeletedAt type: `*time.Time` → `gorm.DeletedAt` (required for v2 soft delete)
- Fixed Base.ID struct tag: `sql:"type:uuid;primary_key"` → `gorm:"type:uuid;primaryKey"` (v2 camelCase)
- Added SQLite WAL mode and busy_timeout pragmas for concurrent read performance
- Replaced `fmt.Println` with `log.Fatalf` on database init failure
- Verified all GORM v2 behavioral changes are accounted for (no CreateTable/DropTable, Count uses int64, all Delete/Update have Where clauses)

## Task Commits

Each task was committed atomically:

1. **Task 1: Upgrade Gin and all remaining non-GORM dependencies** - `9ba3ecf` (feat)
2. **Task 2: Upgrade GORM to v1.26+ with full migration audit** - `3ab8839` (feat)

## Files Created/Modified

- `go.mod` - Upgraded Gin v1.12.0, GORM v1.26.0, zap v1.28.0, godotenv v1.5.1, and all other deps to current versions
- `go.sum` - Updated module graph after dependency upgrades
- `db/base.go` - Changed DeletedAt from `*time.Time` to `gorm.DeletedAt`; changed ID tag from `sql:"type:uuid;primary_key"` to `gorm:"type:uuid;primaryKey"`
- `db/db.go` - Added WAL mode and busy_timeout pragmas; replaced `fmt.Println` with `log.Fatalf`; removed unused `fmt` import

## Decisions Made

- Used `gorm.DeletedAt` instead of `*time.Time` for the Base struct — required for GORM v2 soft delete scope queries (GORM v2 adds `AND deleted_at IS NULL` automatically with this type)
- Used `gorm:"type:uuid;primaryKey"` instead of `sql:"type:uuid;primary_key"` — GORM v2 uses camelCase tag conventions, and `primaryKey` is the correct v2 equivalent
- Used `log.Fatalf` instead of silently returning the DB error — matches the plan's intent per D-13 and BUG-06 (fatal on DB init failure)
- SQLite WAL mode enabled at DB init for better concurrent read performance and to reduce write lock contention
- SQLite busy_timeout=5000ms to handle write contention gracefully

## GORM v2 Migration Audit Results

| Check | Status | Details |
|-------|--------|---------|
| Base.ID tag updated | ✓ | `sql:"type:uuid;primary_key"` → `gorm:"type:uuid;primaryKey"` |
| Base.DeletedAt type updated | ✓ | `*time.Time` → `gorm.DeletedAt` |
| No CreateTable/DropTable calls | ✓ | Only AutoMigrate used |
| All Count() calls use *int64 | ✓ | Already correct — `int64` variables in all 3 call sites |
| All Delete/Update have Where() | ✓ | No bare Delete or Update calls found |
| BeforeCreate hook signature | ✓ | Already used v2 signature `func(tx *gorm.DB) error` |
| No snake_case GORM tags | ✓ | No `auto_increment`, `unique_index`, `polymorphic_value` found |
| WAL mode enabled | ✓ | `PRAGMA journal_mode=WAL` and `PRAGMA busy_timeout=5000` |

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- None. GORM v2 compiled without errors because the existing codebase was already somewhat v2-compatible (hook signature, Count() types, Where() clauses).

## Next Phase Readiness

- All Phase 1 dependencies now at modern versions
- GORM v2 migration complete with all behavioral changes handled
- Ready for Plan 04 (Extract credentials + update Docker config)

---
*Phase: 01-dependency-upgrades*
*Completed: 2026-05-12*
---
phase: 01-dependency-upgrades
plan: 04
subsystem: security,infra
tags: [credentials,docker,env-vars,go-1.24,cgo]

# Dependency graph
requires:
  - phase: 01
    plan: 03
    provides: Dependency upgrades complete (Gin v1.12, GORM v1.26, all deps modernized)
provides:
  - PodcastIndex credentials from environment variables
  - Go 1.24 Docker builder with CGO enabled
  - .env.example documenting all configuration
affects: [deployment,security,search]

# Tech tracking
tech-stack:
  added: [os.Getenv for PodcastIndex credentials, CGO_ENABLED=1 in Dockerfile]
  patterns: [env var config injection, graceful degradation for missing credentials]

key-files:
  created:
    - .env.example
  modified:
    - service/itunesService.go
    - Dockerfile
    - docker-compose.yml

key-decisions:
  - "PodcastIndex credentials loaded from env vars with warning + nil return when missing (graceful, not fatal)"
  - "Replaced log.Fatal with fmt.Println for PodcastIndex search errors — no process crash on API failure"
  - "Go 1.24 builder with CGO_ENABLED=1 and gcc/musl-dev for SQLite CGO compilation"
  - "Docker-compose uses placeholder values for PodcastIndex env vars"

patterns-established:
  - "Sensitive credentials read from environment, not hardcoded in source"
  - "Missing optional credentials produce warnings, not fatal errors"

requirements-completed: [DEPS-09, DEPS-10]

# Metrics
duration: 5min
completed: 2026-05-12
---

# Phase 1 Plan 04: Extract Hardcoded Credentials + Update Docker Configuration Summary

**PodcastIndex credentials moved to env vars; Docker updated for Go 1.24+ with CGO; .env.example documents all config**

## Performance

- **Duration:** 5 min
- **Started:** 2026-05-12T09:53:27Z
- **Completed:** 2026-05-12T09:58:36Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Removed hardcoded PodcastIndex API key (LNGTNUAFVL9W2AQKVZ49) and secret from source code
- Replaced with `os.Getenv()` lookups following existing env var pattern (PASSWORD, CONFIG, DATA, CHECK_FREQUENCY)
- Added graceful degradation: warning log when credentials missing, returns nil (search disabled)
- Replaced `log.Fatal()` with `fmt.Println()` for PodcastIndex search errors — no more process crash on API failure
- Created `.env.example` documenting all environment variables (CONFIG, DATA, CHECK_FREQUENCY, PASSWORD, PODCASTINDEX_KEY, PODCASTINDEX_SECRET, PUID, PGID)
- Updated Dockerfile from Go 1.15.2 to Go 1.24 builder image
- Added `gcc` and `musl-dev` to Alpine packages for CGO/SQLite compilation
- Added `CGO_ENABLED=1` to `go build` command in Dockerfile
- Added `PODCASTINDEX_KEY` and `PODCASTINDEX_SECRET` to `docker-compose.yml`
- Preserved `/config` and `/assets` volume mount pattern in both files

## Task Commits

1. **Task 1: Extract PodcastIndex credentials to environment variables** - `7030421` (feat)
2. **Task 2: Update Dockerfile and docker-compose.yml for Go 1.24+** - `312da38` (feat)

## Files Created/Modified

- `service/itunesService.go` - Removed hardcoded `const` block; replaced with `os.Getenv()` lookups; removed `log` import, added `os` import; replaced `log.Fatal` with `fmt.Println` warning
- `.env.example` - New file documenting all 7 environment variables
- `Dockerfile` - Updated Go 1.15.2→1.24; added gcc/musl-dev; added CGO_ENABLED=1
- `docker-compose.yml` - Added PODCASTINDEX_KEY and PODCASTINDEX_SECRET environment variables

## Decisions Made

- Used `fmt.Println` warning + `nil` return when PodcastIndex credentials are missing rather than fatal exit — search is additive functionality, not required for core download loop
- Replaced `log.Fatal(err.Error())` with `fmt.Println()` for search errors — matching existing codebase error handling patterns and avoiding process crash on transient API errors
- Docker-compose uses placeholder values (`your_key_here`, `your_secret_here`) — users must provide their own credentials
- `.env.example` serves as both documentation and a template that can be copied to `.env` for local development

## Deviations from Plan

None - plan executed exactly as written.

## Threat Model Compliance

| Threat | Disposition | Status |
|--------|-------------|--------|
| T-04-01 (I - Hardcoded credentials) | mitigate | ✓ Credentials removed from source; loaded via env vars |
| T-04-02 (E - Missing env vars) | mitigate | ✓ Warning logged, search gracefully disabled |
| T-04-03 (S - Docker image) | accept | ✓ Alpine + CGO is standard; no new attack surface |

## Verification Results

1. ✅ `service/itunesService.go` contains `os.Getenv("PODCASTINDEX_KEY")` and `os.Getenv("PODCASTINDEX_SECRET")`
2. ✅ `service/itunesService.go` does NOT contain hardcoded API key `LNGTNUAFVL9W2AQKVZ49`
3. ✅ `.env.example` exists with `PODCASTINDEX_KEY` and `PODCASTINDEX_SECRET` entries
4. ✅ `Dockerfile` uses `golang:1.24-alpine` builder
5. ✅ `Dockerfile` has `CGO_ENABLED=1` in build command
6. ✅ `docker-compose.yml` includes `PODCASTINDEX_KEY` and `PODCASTINDEX_SECRET`
7. ✅ `go build ./...` compiles successfully
8. ✅ Volume mount pattern (`/config`, `/assets`) preserved in Docker configuration

---
*Phase: 01-dependency-upgrades*
*Completed: 2026-05-12*
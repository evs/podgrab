# Technology Stack

**Analysis Date:** 2026-05-11

## Languages

**Primary:**
- Go 1.15 — Backend server, all business logic, controllers, models, services

**Secondary:**
- HTML/JavaScript/CSS — Frontend templates served via Go's `html/template` (server-side rendered)
  - Frontend uses Vue.js 2, Axios, Luxon, Amplitude.js, Tippy.js, vue-multiselect, vue-toasted
  - Located in `client/` (templates) and `webassets/` (static assets)

## Runtime

**Environment:**
- Go runtime (compiled binary)
- Docker container (Alpine Linux base image)
- `GIN_MODE=release` for production

**Package Manager:**
- Go Modules — `go.mod` present
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Gin v1.7.2 — HTTP web framework (`github.com/gin-gonic/gin`)
- Gin Contrib Location v0.0.2 — URL/location helper (`github.com/gin-contrib/location`)
- Gorilla WebSocket v1.4.2 — WebSocket support for real-time player communication

**ORM / Data:**
- GORM v1.20.2 — ORM for SQLite (`gorm.io/gorm`)
- GORM SQLite Driver v1.1.3 — SQLite dialect (`gorm.io/driver/sqlite`)

**Testing:**
- Not detected — no test files found in the repository

**Build/Dev:**
- Docker — multi-stage build (Go builder → Alpine runtime)
- Docker Compose v2.1 — container orchestration

## Key Dependencies

**Critical:**
- `github.com/TheHippo/podcastindex` v1.0.0 — Podcast Index API client for podcast search
- `github.com/antchfx/xmlquery` v1.3.3 — XML DOM querying (RSS feed image extraction)
- `github.com/dgrijalva/jwt-go` v3.2.0+incompatible — JWT token support (unused or future auth)
- `golang.org/x/crypto` — Cryptographic functions ( bcrypt password hashing)
- `go.uber.org/zap` v1.16.0 — Structured logging

**Infrastructure:**
- `github.com/joho/godotenv` v1.3.0 — `.env` file loading (autoloaded via `_ "github.com/joho/godotenv/autoload"`)
- `github.com/jasonlvhit/gocron` v0.0.1 — Job scheduling (episode refresh, file checks, backups)
- `github.com/satori/go.uuid` v1.2.0 — UUID generation for model primary keys
- `github.com/gobeam/stringy` v0.0.0-20200717095810-8a3637503f62 — String manipulation (kebab-case for filenames)
- `github.com/grokify/html-strip-tags-go` v0.0.0-20200923094847-079d207a09f1 — HTML tag stripping
- `github.com/microcosm-cc/bluemonday` v1.0.15 — HTML sanitization policy
- `github.com/chris-ramon/douceur` v0.2.0 — CSS sanitization (indirect dependency)

## Configuration

**Environment:**
- `.env` file via `godotenv/autoload` — contains secrets/config (never read in code)
- `CONFIG` env var — Path to configuration directory (database, backups)
- `DATA` env var — Path to assets/data directory (downloaded podcast files)
- `PASSWORD` env var — Optional HTTP Basic Auth credentials (user: "podgrab")
- `CHECK_FREQUENCY` env var — Cron check interval in minutes (default: 30)
- `PUID` / `PGID` env var — File ownership UID/GID for downloaded files
- `GIN_MODE` env var — Gin framework mode (`release` for production)

**Build:**
- `Dockerfile` — Multi-stage Go build, Alpine runtime
- `docker-compose.yml` — Service orchestration with volume mounts

## Platform Requirements

**Development:**
- Go 1.15+ (specified in `go.mod`)
- SQLite3 (via CGO — `gorm.io/driver/sqlite` requires `gcc` and `CGO_ENABLED=1`)

**Production:**
- Docker (Alpine-based image `akhilrex/podgrab`)
- Persistent volumes: `/config` (database, backups), `/assets` (downloaded media)
- Port 8080 (default, hardcoded in `main.go` via `r.Run()`)

---

*Stack analysis: 2026-05-11*
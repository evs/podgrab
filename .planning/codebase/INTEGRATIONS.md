# External Integrations

**Analysis Date:** 2026-05-11

## APIs & External Services

**Podcast Search — Apple iTunes:**
- Service: iTunes Search API
- Purpose: Podcast discovery and search (`service/itunesService.go`)
- Endpoint: `https://itunes.apple.com/search?term={q}&entity=podcast`
- SDK/Client: Raw HTTP via `net/http` — `makeQuery()` in `service/podcastService.go`
- Auth: None (public API)

**Podcast Search — Podcast Index:**
- Service: Podcast Index API
- Purpose: Alternative podcast search (`service/itunesService.go`)
- SDK/Client: `github.com/TheHippo/podcastindex` v1.0.0
- Auth: API key + secret (hardcoded in `service/itunesService.go:42-44`)
  - Key: `PODCASTINDEX_KEY = "LNGTNUAFVL9W2AQKVZ49"`
  - Secret: `PODCASTINDEX_SECRET = "H8tq^CZWYmAywbnngTwB$rwQHwMSR8#fJb#Bhgb3"`

**Podcast Search — gpodder.net:**
- Service: gpodder.net open podcast directory
- Purpose: Podcast search, tag browsing, top lists (`service/gpodderService.go`)
- Endpoints:
  - Search: `https://gpodder.net/search.json?q={q}`
  - By tag: `https://gpodder.net/api/2/tag/{tag}/{count}.json`
  - Top list: `https://gpodder.net/toplist/{count}.json`
  - Tags: `https://gpodder.net/api/2/tags/{count}.json`
- Auth: None (public API)

**RSS Feed Fetching:**
- Purpose: Podcast feed parsing and episode data extraction (`service/podcastService.go`)
- Mechanism: Raw HTTP GET to podcast RSS URLs → XML unmarshal into `model.PodcastData`
- Additional: `xmlquery` for extracting iTunes-specific image URLs from feeds

## Data Storage

**Databases:**
- SQLite — Embedded file-based database
  - Location: `{CONFIG}/podgrab.db` (configured via `CONFIG` env var)
  - Client/ORM: GORM v1.20.2 (`gorm.io/gorm` + `gorm.io/driver/sqlite`)
  - Schema models: `db/podcast.go` — Podcast, PodcastItem, Setting, Migration, JobLock, Tag
  - Migrations: Auto-migrate via `db.AutoMigrate()` + manual migrations in `db/migrations.go`
  - Connection pool: Max 10 idle connections (`db/db.go:30`)
  - Uses CGO (requires GCC for compilation)

**File Storage:**
- Local filesystem — Downloaded podcast episodes and images
  - Media files: `{DATA}/{podcast_name}/{filename}` (configured via `DATA` env var)
  - Cover images: `{DATA}/{podcast_name}/folder.jpg`
  - Episode images: `{DATA}/{podcast_name}/images/{filename}.jpg`
  - NFO files: `{DATA}/{podcast_name}/album.nfo` (optional, setting-driven)
  - Backups: `{CONFIG}/backups/podgrab_backup_{timestamp}.tar.gz` (database only, rotates keeping 5)

**Caching:**
- None — No caching layer detected

## Authentication & Identity

**Auth Provider:**
- Custom — HTTP Basic Auth via Gin middleware
  - Implementation: `main.go:135-143`
  - Conditional: Only active when `PASSWORD` env var is set
  - Single user: username "podgrab", password from `PASSWORD` env var
  - If `PASSWORD` is empty, no authentication is applied
  - Library: `gin.BasicAuth` from `gin-gonic/gin`

**JWT:**
- `github.com/dgrijalva/jwt-go` is listed as a dependency but not observed in active use in the codebase

## Monitoring & Observability

**Error Tracking:**
- None — No Sentry, Rollbar, or similar error tracking service detected

**Logging:**
- `go.uber.org/zap` — Structured JSON logging (`service/podcastService.go:24-29`)
  - Used primarily in `service/podcastService.go` and `service/fileService.go` via `Logger.Errorw()`
  - Production logger (JSON format) initialized in `init()`
- `fmt.Println` — Scattered debug logging throughout controllers and services
- `log.Print` / `log.Fatal` — Standard library logging in cron and startup

## CI/CD & Deployment

**Hosting:**
- Docker Hub image: `akhilrex/podgrab`
- Self-hosted via Docker Compose

**CI Pipeline:**
- GitHub Actions workflow exists at `.github/workflows/hub.yml`
  - Likely a Docker Hub build/push workflow (not fully inspected)

## Environment Configuration

**Required env vars:**
- `CONFIG` — Path to config directory (database, backups), defaults to `/config` in Docker
- `DATA` — Path to data directory (downloaded media), defaults to `/assets` in Docker
- `CHECK_FREQUENCY` — Cron interval in minutes, default 30

**Optional env vars:**
- `PASSWORD` — Enable HTTP Basic Auth with username "podgrab"
- `PUID` / `PGID` — File ownership UID/GID for downloaded files
- `GIN_MODE` — Set to `release` in Docker for production mode

**Secrets location:**
- `.env` file present — Contains environment configuration (existence noted, contents not read)
- Podcast Index API credentials hardcoded in `service/itunesService.go:42-44`

## Webhooks & Callbacks

**Incoming:**
- WebSocket endpoint at `/ws` — Real-time player communication (`controllers/websockets.go`)
  - Supports message types: `RegisterPlayer`, `PlayerRemoved`, `Enqueue`, `Register`
  - No external webhooks or API callback endpoints

**Outgoing:**
- None — No outgoing webhooks or callback registrations

## RSS Feed Generation

**Outgoing RSS:**
- Podgrab generates RSS feeds for consumed podcasts (`controllers/podcast.go:455-563`)
  - `/podcasts/{id}/rss` — RSS feed for a specific podcast
  - `/tags/{id}/rss` — RSS feed for episodes with a specific tag
  - `/rss` — Global RSS feed for all episodes
  - Models: `model/rssModels.go`
  - Uses `model.RssPodcastData` structure with iTunes, Media, Atom, PSC namespaces

---

*Integration audit: 2026-05-11*
# Architecture Patterns

**Domain:** Self-hosted podcast manager — Go monolith with server-rendered UI being incrementally modernized to add REST API + React SPA
**Researched:** 2026-05-12

## Recommended Architecture: Strangler Fig with Dual-Route Gateway

The core strategy is the **Strangler Fig pattern**: the React SPA gradually replaces server-rendered pages, one route at a time, while the existing Go server continues serving both old and new UIs simultaneously. The Go server acts as a unified gateway that routes to either the legacy SSR pages or the React SPA based on URL prefix.

```text
┌─────────────────────────────────────────────────────────────────────┐
│                         Go Server (Gin)                             │
│                                                                     │
│  ┌─────────────────────┐  ┌──────────────────────────────────────┐  │
│  │ /api/v1/*           │  │ /* (catch-all)                       │  │
│  │ REST API handlers   │  │ Route decision middleware            │  │
│  │ → service layer     │  │  ├─ SSR page routes (legacy)         │  │
│  │ → db layer          │  │  │   /add, /, /podcasts/:id/view    │  │
│  │                     │  │  │   /episodes, /allTags, /settings  │  │
│  │                     │  │  │   /backups, /player               │  │
│  │                     │  │  └─ React SPA routes (new)           │  │
│  │                     │  │      /app/* → serve index.html       │  │
│  │                     │  │      React Router handles the rest   │  │
│  └─────────────────────┘  └──────────────────────────────────────┘  │
│                                                                     │
│  ┌─────────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │ /ws             │  │ /assets/*    │  │ /web/* (new SPA build) │  │
│  │ WebSocket       │  │ File serving │  │ Static React bundle    │  │
│  │ Player sync     │  │ (downloads)  │  │ JS/CSS from Vite build │  │
│  └─────────────────┘  └──────────────┘  └────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
         │                    │                       │
         ▼                    ▼                       ▼
┌─────────────────┐  ┌──────────────┐  ┌────────────────────────────┐
│  Service Layer   │  │  File System │  │  SQLite (GORM)             │
│  (shared by both │  │  Downloads   │  │  podgrab.db                │
│   SSR + API)     │  │  Backups     │  │                            │
└─────────────────┘  └──────────────┘  └────────────────────────────┘
```

## Component Boundaries

### Component Map

| Component | Responsibility | Communicates With | New or Existing |
|-----------|---------------|-------------------|-----------------|
| **API Router** (`/api/v1/*`) | REST JSON endpoints for React SPA | Service layer, db layer | **NEW** |
| **Page Router** (legacy `/*`) | Server-rendered HTML pages (Go templates + Vue 2) | Service layer, db layer | Existing |
| **SPA Fallback** (`/app/*`) | Serve React SPA `index.html`, let React Router handle client-side routing | Static file server only | **NEW** |
| **React SPA** | Client-side UI: podcast list, episode list, player, settings, tags | API Router (`/api/v1/*`), WebSocket (`/ws`) | **NEW** |
| **WebSocket Hub** | Real-time player state sync across tabs/clients | Both old and new UI (same protocol) | Existing (needs mutex fix) |
| **Service Layer** | Business logic: podcast CRUD, download, search, OPML | DB layer, file system, external APIs | Existing (refactored incrementally) |
| **DB Layer** | GORM models, CRUD queries, migrations | SQLite | Existing (needs DI refactor) |
| **Auth Middleware** | HTTP Basic Auth on all routes (both old and new) | Gin middleware chain | Existing (shared) |
| **Cron Scheduler** | Background episode refresh, file checks, backups | Service layer | Existing |

### Boundary Rules

1. **API handlers MUST go through service layer** — never call `db.*` directly. Current controller bypasses (`GetAllTags`, `AddTagToPodcast`, etc.) must be routed through service functions first.
2. **Service layer is shared** — both old page controllers and new API handlers call the same `service.*` functions. This ensures data consistency during migration.
3. **React SPA NEVER talks to SSR page routes** — the SPA only communicates via `/api/v1/*` REST endpoints and `/ws` WebSocket.
4. **Old UI and new UI share auth** — the same Basic Auth middleware protects both `/api/v1/*` and `/*`. The SPA passes credentials via `Authorization` header.
5. **WebSocket protocol is unchanged** — both old Vue 2 player and new React player use the same WebSocket message format (`RegisterPlayer`, `Enqueue`, etc.).

## Data Flow

### Current Flow (Server-Rendered)

```text
Browser → GET /podcasts/:id/view → Page Controller → service.* + db.*
                                              ↓
                                    Render HTML template with data
                                              ↓
                                    Browser receives complete HTML
                                    Inline Vue 2 makes AJAX calls to existing
                                    JSON endpoints for search, episode ops
```

### New Flow (React SPA)

```text
Browser → GET /app/podcasts/:id → Gin serves /app/index.html
                                      ↓
                              React bootstraps, Router matches /podcasts/:id
                                      ↓
                              React component fetches GET /api/v1/podcasts/:id
                                      ↓
                              API handler → service.* + db.* → JSON response
                                      ↓
                              React renders component with data
```

### Coexistence Flow

```text
Browser
  ├── Old bookmark: /podcasts/abc123/view  → SSR page (existing, unchanged)
  ├── SPA navigation: /app/podcasts/abc123  → React SPA (new)
  └── API call: /api/v1/podcasts/abc123     → JSON (consumed by SPA only)
```

Both flows use the same service layer and database. The same podcast data powers both the SSR template and the JSON response.

### Shared WebSocket Flow (Unchanged)

```text
React Player ──── WebSocket /ws ──── Go WebSocket Hub ──── Other tabs (old or new UI)
                  (same message format for both)
```

## Patterns to Follow

### Pattern 1: API Version Prefix

**What:** All new REST endpoints live under `/api/v1/`, separate from existing routes
**When:** Creating any new JSON endpoint for the React SPA
**Example:**
```go
// main.go - route registration
apiV1 := router.Group("/api/v1")
{
    apiV1.GET("/podcasts", api.GetAllPodcasts)
    apiV1.GET("/podcasts/:id", api.GetPodcastByID)
    apiV1.POST("/podcasts", api.AddPodcast)
    // ...
}
```
**Why:** Prevents collision with existing routes (`GET /podcasts` already returns JSON but with inconsistent patterns). The `/api/v1/` prefix gives a clean contractual boundary, allows the API to evolve independently, and makes it trivial to see which routes are "new API" vs "legacy".

### Pattern 2: API DTO Layer (Response Shaping)

**What:** API handlers return purpose-built DTOs, not raw GORM models
**When:** Any `/api/v1/*` endpoint
**Example:**
```go
// api/dto/podcast.go
type PodcastDTO struct {
    ID                string   `json:"id"`
    Title             string   `json:"title"`
    URL               string   `json:"url"`
    Image             string   `json:"image"`
    IsPaused          bool     `json:"isPaused"`
    DownloadedCount   int      `json:"downloadedCount"`
    TotalCount        int      `json:"totalCount"`
    Tags              []TagDTO `json:"tags"`
    LastEpisodeDate   string   `json:"lastEpisodeDate,omitempty"`
}

// api/handlers/podcast.go
func GetPodcastByID(c *gin.Context) {
    podcast := service.GetPodcastByID(id)
    dto := dto.FromPodcast(podcast)
    c.JSON(200, dto)
}
```
**Why:** GORM models (`db.Podcast`) leak database structure (GORM tags, `gorm:"-"` computed fields, association loading behavior). A DTO layer decouples the API contract from the DB schema, lets you shape responses for the SPA's needs, and avoids accidental exposure of internal fields.

### Pattern 3: Service Layer Must Be the Only Gateway

**What:** All API handlers call `service.*` functions, never `db.*` directly
**When:** Any new API handler, and refactoring existing page controllers
**Current violation examples:**
```go
// BAD: controllers/podcast.go:425 — controller directly calls db
tags, err := db.GetAllTags("")

// BAD: controllers/podcast.go:601 — controller directly calls db
err := db.AddTagToPodcast(addRemoveTagQuery.Id, addRemoveTagQuery.TagId)

// GOOD: controllers/podcast.go:406 — controller goes through service
pod, err := service.AddPodcast(addPodcastData.Url)
```
**Fix:** Before building the API layer, route the 6 direct `db.*` calls in controllers through service-layer wrappers:
- `db.GetAllTags` → `service.GetAllTags`
- `db.GetTagById` → `service.GetTagByID`
- `db.AddTagToPodcast` → `service.AddTagToPodcast`
- `db.RemoveTagFromPodcast` → `service.RemoveTagFromPodcast`
- `db.GetPaginatedPodcastItemsNew` → `service.GetFilteredEpisodes`
- `db.GetPodcastItemById` → (already wrapped in some places, not all)

### Pattern 4: SPA Route Prefix `/app/*`

**What:** React SPA is served from `/app/*` route prefix; Gin catches all `/app/*` routes and serves the SPA's `index.html`
**When:** Serving the React SPA to the browser
**Example:**
```go
// main.go
// Serve React SPA static build
router.Static("/web", "./spa/dist")   // JS/CSS bundles
router.NoRoute(func(c *gin.Context) {
    // Fallback: if path starts with /app, serve SPA index.html
    if strings.HasPrefix(c.Request.URL.Path, "/app") {
        c.File("./spa/dist/index.html")
        return
    }
    // Otherwise, 404
    c.JSON(404, gin.H{"error": "Not found"})
})
```
**Why:** The `/app/*` prefix creates a clean namespace where the SPA "owns" all routes. React Router handles `/app/podcasts`, `/app/podcasts/:id`, `/app/settings`, etc. client-side. No collision with legacy SSR routes (`/`, `/add`, `/podcasts/:id/view`). Users can bookmark SPA URLs. The prefix makes it obvious which UI a user is looking at.

### Pattern 5: Gradual Service Extraction

**What:** Extract service methods slowly as API endpoints need them, rather than refactoring the entire service layer upfront
**When:** Building each new API endpoint
**Example:**
```go
// Phase 1: Just wrap existing calls
func (s *PodcastService) GetByID(id string) (*db.Podcast, error) {
    var podcast db.Podcast
    err := db.GetPodcastById(id, &podcast)
    return &podcast, err
}

// Phase 2 (later): Inject *gorm.DB instead of global
func (s *PodcastService) GetByID(ctx context.Context, id string) (*Podcast, error) {
    var podcast Podcast
    err := s.db.WithContext(ctx).First(&podcast, "id = ?", id).Error
    return &podcast, err
}
```
**Why:** Big-bang DI refactoring blocks all progress. Wrapping existing `db.*` calls in service methods is a fast, low-risk step that unblocks API development immediately. The DI refactor can happen in a later pass once the API is working.

### Pattern 6: Embedded SPA with Go's `embed`

**What:** Use Go's `//go:embed` directive to embed the React build output directly into the Go binary
**When:** Building Docker images and distributing the application
**Example:**
```go
// spa/embed.go
//go:embed dist/*
var SpaAssets embed.FS

// main.go
spaFS, _ := fs.Sub(SpaAssets, "dist")
router.StaticFS("/web", http.FS(spaFS))
```
**Why:** Eliminates the need for a separate static file server or `COPY client` step in Docker. The Go binary becomes self-contained (templates + API + SPA all in one binary). This matches the existing deployment model (single Alpine binary) and avoids CORS issues since everything is same-origin.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Rewrite Instead of Strangle

**What:** Replacing the entire server-rendered UI in one big-bang cutover
**Why bad:** Violates the project constraint of "incremental migration, no flag-day cutover". If the SPA is incomplete at cutover, users lose features. If you discover a bug, you can't easily roll back.
**Instead:** Use the Strangler Fig pattern. Replace one page at a time. Both UIs coexist. When a React view is feature-complete and tested, redirect the old SSR route to the `/app/*` equivalent. Only remove the old route when you're confident the SPA replacement is solid.

### Anti-Pattern 2: API Endpoints on Existing Routes

**What:** Adding API behavior to existing routes (e.g., making `GET /podcasts` return "better JSON" for the SPA)
**Why bad:** The existing `GET /podcasts` already returns JSON but with Podgrab-specific conventions (no pagination metadata envelope, `gorm:"-"` computed fields in response, inconsistent error format). Modifying it risks breaking the existing Vue 2 code that consumes it.
**Instead:** New API endpoints live on `/api/v1/*`. The existing routes remain untouched until their SSR page is fully replaced. Once a page is replaced, the old route can redirect to the SPA.

### Anti-Pattern 3: Sharing GORM Models as API Responses

**What:** Returning `db.Podcast` directly as JSON from API handlers (which is what the existing controllers do)
**Why bad:** GORM models include `gorm:"-"` fields that aren't populated consistently, associations that may or may not be preloaded (so they appear as `null` or empty), and internal fields like `DownloadStatus int` that need translation for a front-end consumer. The current `GetAllPodcasts` already returns raw GORM models — it works but is fragile.
**Instead:** API DTOs with explicit JSON tags, computed/derived fields resolved properly, and null-safe defaults. The SPA gets a clean, documented contract.

### Anti-Pattern 4: Dual WebSocket Protocols

**What:** Creating a separate WebSocket endpoint or message protocol for the React SPA
**Why bad:** Player sync must work across tabs regardless of whether the tab is running old Vue 2 UI or new React UI. Two protocols means they can't talk to each other.
**Instead:** Reuse the exact same `/ws` endpoint and message format. The React player sends/receives the same `Message` structs (`RegisterPlayer`, `Enqueue`, `PlayerExists`, `NoPlayer`). Fix the mutex issue on the WebSocket maps, but don't change the protocol.

### Anti-Pattern 5: API-First Bulk Design

**What:** Designing and building ALL API endpoints before building ANY React components
**Why bad:** You'll over-design APIs you don't need yet, miss APIs you do need, and delay getting visual feedback. The project explicitly decided "API + frontend built together, add API endpoints as React components need them, not all upfront."
**Instead:** Build the React SPA one view at a time. When a component needs data, add the API endpoint it needs. This ensures every API endpoint is actually needed, and you get UI feedback immediately.

### Anti-Pattern 6: Separate Frontend Dev Server in Production

**What:** Running Vite dev server as a separate process alongside Go in production (CORS, different ports)
**Why bad:** Complicates deployment, introduces CORS issues, breaks single-binary Docker model, adds a second process to manage
**Instead:** Vite dev server is for development only. In production, `npm run build` creates static files in `spa/dist/`, Go embeds them via `//go:embed`, and Gin serves them. One process, one binary, one port — just like today.

## Coexistence Strategy: Page-by-Page Strangulation

### Migration Sequence

Each SSR page is replaced independently. A page is "done" when its React equivalent handles all the same user actions and the SSR route redirects to the SPA route.

```text
Phase A: Foundation (no UI changes yet)
├── Add /api/v1/* route group
├── Create api/handlers/ package
├── Create api/dto/ package  
├── Create SPA scaffold (Vite + React + Router)
├── Fix service layer bypasses (6 db.* calls in controllers)
├── Fix WebSocket mutex
└── Set up //go:embed for SPA

Phase B: First SPA Page (Podcast List — /app/)
├── API: GET /api/v1/podcasts (list with stats)
├── API: DELETE /api/v1/podcasts/:id
├── API: GET /api/v1/podcasts/:id/pause, /unpause
├── React: PodcastList component
└── Redirect: GET / → GET /app/

Phase C: Podcast Detail + Episodes (/app/podcasts/:id)
├── API: GET /api/v1/podcasts/:id (with items, paginated)
├── API: GET /api/v1/podcasts/:id/items (paginated, filtered)
├── API: GET /api/v1/podcastitems/:id/download
├── API: DELETE /api/v1/podcastitems/:id
├── React: PodcastDetail, EpisodeList components
└── Redirect: GET /podcasts/:id/view → /app/podcasts/:id

Phase D: Add Podcast + Search (/app/add)
├── API: POST /api/v1/podcasts (add by URL)
├── API: GET /api/v1/search?q=&type=
├── React: AddPodcast, SearchResults components
└── Redirect: GET /add → /app/add

Phase E: Tags (/app/tags)
├── API: GET /api/v1/tags
├── API: POST /api/v1/tags, DELETE /api/v1/tags/:id
├── API: POST /api/v1/podcasts/:id/tags/:tagId
├── API: DELETE /api/v1/podcasts/:id/tags/:tagId
├── React: TagsPage component
└── Redirect: GET /allTags → /app/tags

Phase F: Player (/app/player)
├── API: GET /api/v1/podcastitems (filtered, for enqueue)
├── API: GET /api/v1/podcastitems/:id/file
├── React: Player component (with Howler.js or HTML5 Audio)
├── Reuse: Same /ws WebSocket connection
└── Redirect: GET /player → /app/player

Phase G: Settings + Backups + OPML (/app/settings, /app/backups)
├── API: GET /api/v1/settings
├── API: PUT /api/v1/settings  
├── API: GET /api/v1/backups
├── API: POST /api/v1/opml (import)
├── API: GET /api/v1/opml (export)
├── React: Settings, Backups components
└── Redirect: GET /settings → /app/settings, GET /backups → /app/backups

Phase H: Cleanup (all SSR pages replaced)
├── Remove /client/*.html templates
├── Remove /webassets/ vendored libraries
├── Remove template FuncMap from main.go
├── Make /app/* the default route
└── Remove legacy page controllers
```

### Redirect Strategy

When an SSR page is fully replaced, add a redirect from the old route to the new SPA route:

```go
// After Phase B completes:
router.GET("/", func(c *gin.Context) {
    c.Redirect(302, "/app/")
})
```

During the transition period, old routes and new routes both work. Users with old bookmarks land on the SSR page. Users who navigate from the SPA stay in the SPA. No forced migration until each page is ready.

### API Endpoint Mapping: Existing → New

| Existing Route | Existing Returns | New API Route | SPA Route | Status |
|---|---|---|---|---|
| `GET /podcasts` | JSON (raw GORM) | `GET /api/v1/podcasts` | `/app/` | Phase B |
| `POST /podcasts` | JSON | `POST /api/v1/podcasts` | `/app/add` | Phase D |
| `GET /podcasts/:id` | JSON (raw GORM, `db.GetPodcastById`) | `GET /api/v1/podcasts/:id` | `/app/podcasts/:id` | Phase C |
| `DELETE /podcasts/:id` | 204 | `DELETE /api/v1/podcasts/:id` | — | Phase B |
| `GET /podcasts/:id/items` | JSON (raw GORM, `db.GetAllPodcastItemsByPodcastId`) | `GET /api/v1/podcasts/:id/items` | `/app/podcasts/:id` | Phase C |
| `GET /podcasts/:id/pause` | JSON | `POST /api/v1/podcasts/:id/pause` | — | Phase B |
| `GET /podcasts/:id/unpause` | JSON | `POST /api/v1/podcasts/:id/unpause` | — | Phase B |
| `GET /podcastitems` | JSON (paginated) | `GET /api/v1/episodes` | `/app/episodes` | Phase C |
| `PATCH /podcastitems/:id` | JSON | `PATCH /api/v1/episodes/:id` | — | Phase C |
| `GET /podcastitems/:id/download` | JSON | `POST /api/v1/episodes/:id/download` | — | Phase C |
| `GET /podcastitems/:id/markPlayed` | — | `POST /api/v1/episodes/:id/played` | — | Phase C |
| `GET /podcastitems/:id/markUnplayed` | — | `POST /api/v1/episodes/:id/unplayed` | — | Phase C |
| `GET /tags` | JSON | `GET /api/v1/tags` | `/app/tags` | Phase E |
| `POST /tags` | JSON | `POST /api/v1/tags` | — | Phase E |
| `GET /search` | JSON | `GET /api/v1/search` | `/app/add` | Phase D |
| `GET /` | HTML | redirect → `/app/` | `/app/` | Phase B |
| `GET /add` | HTML | — | `/app/add` | Phase D |
| `GET /podcasts/:id/view` | HTML | — | `/app/podcast/:id` | Phase C |
| `GET /episodes` | HTML | — | `/app/episodes` | Phase C |
| `GET /allTags` | HTML | — | `/app/tags` | Phase E |
| `GET /settings` | HTML | `GET /api/v1/settings` | `/app/settings` | Phase G |
| `GET /backups` | HTML | `GET /api/v1/backups` | `/app/backups` | Phase G |
| `GET /player` | HTML | — | `/app/player` | Phase F |

**Key improvements in new API routes:**
- **GET→POST for mutations**: Old routes use `GET` for state-changing operations (`/podcasts/:id/pause`, `/podcastitems/:id/markPlayed`). New API uses proper HTTP methods.
- **Consistent resource naming**: `/api/v1/episodes` (not `/podcastitems`) — clearer for the React frontend.
- **DTO responses**: All new endpoints return properly shaped DTOs with consistent field names (`snake_case` JSON, not a mix of PascalCase and camelCase).

## Build Order (Component Dependencies)

```text
1. Fix service layer bypasses          ← unblocks clean API handlers
   ↓
2. Set up /api/v1/ route group + DTOs ← unblocks React components
   ↓
3. Set up SPA scaffold (Vite+React)  ← unblocks UI development
   ↓                                    (can parallel with step 2)
4. Set up //go:embed + SPA serving    ← unblocks development workflow
   ↓
5. Build API+SPA per page            ← depends on 1-4, done per-page
   ↓                                  (Phase B through G above)
6. Cleanup legacy SSR code           ← depends on all pages migrated
```

**Parallelization opportunities:**
- Steps 2 and 3 can happen in parallel
- After step 5, each page (B through G) can be developed somewhat independently
- The player (Phase F) is the most complex due to WebSocket + audio; it should come after the team is comfortable with the SPA pattern from earlier pages

## Scalability Considerations

| Concern | Current (100 users) | Migration (1K SPA requests) | Future (10K podcast library) |
|---------|----------------------|----------------------------|-------------------------------|
| SQLite concurrency | Fine for single-user | Same — still single-user self-hosted | Enable WAL mode; add retry logic |
| API response time | N/A (SSR = ~same) | Negligible overhead (JSON vs HTML) | Add pagination to all list endpoints |
| SPA bundle size | N/A | ~200KB gzipped (React + Router) | Code-split by route (Vite does this) |
| Memory (Go binary) | ~30MB | +5-10MB for SPA embed | No significant growth from embed |
| WebSocket connections | ~5-10 tabs | Same | Sync.Map + mutex fixes matter more |

## Existing API Endpoints That Already Return JSON

An important discovery: many existing routes already return JSON, not HTML. This means the existing controllers already do much of what the API layer needs — but with problems:

| Existing JSON Route | Problem | API Fix |
|---|---|---|
| `GET /podcasts` | Returns raw `[]db.Podcast` with inconsistent `gorm:"-"` fields | DTO with computed stats |
| `GET /podcasts/:id` | Calls `db.GetPodcastById` directly, no service wrap | Route through service, return DTO |
| `GET /podcastitems` | Returns raw GORM + filter metadata | Clean paginated response with DTOs |
| `GET /podcastitems/:id` | Calls `db.GetPodcastItemById` directly | Route through service, return DTO |
| `DELETE /podcasts/:id` | Returns `204` with empty body — fine | Keep, or return `200` with deleted resource |
| `GET /podcasts/:id/pause` | Uses GET for mutation, no request body | Change to POST |
| `GET /podcastitems/:id/markPlayed` | Uses GET for mutation | Change to POST |

**Implication for build order:** The existing JSON routes mean we're not building entirely from scratch. The migration path is: (1) create `/api/v1/` versions that fix the HTTP method and response shape issues, (2) don't modify existing routes until the SPA is using the new ones exclusively, (3) then deprecate old routes.

## CORS Strategy

Since the SPA is served from the same origin (embedded in Go binary, same port 8080), **CORS is not needed for production**. The React SPA makes requests to `/api/v1/*` on the same host:port. No preflight, no CORS headers needed.

For **development only** (Vite dev server on port 5173, Go on port 8080), add CORS middleware:

```go
// main.go — development only
if os.Getenv("GIN_MODE") != "release" {
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
    }))
}
```

## Directory Structure (Target)

```text
podgrab/
├── main.go                    # Router setup, middleware, SPA serving
├── db/                        # Existing — GORM models, CRUD, migrations
├── model/                     # Existing — DTOs for external data (RSS, OPML, iTunes)
├── service/                   # Existing — business logic (refactored to close bypasses)
├── controller/                # Existing — SSR page controllers (removed in Phase H)
├── api/                       # NEW — REST API layer
│   ├── handler/               # NEW — Gin handler functions for /api/v1/*
│   │   ├── podcast.go         # Podcast CRUD endpoints
│   │   ├── episode.go         # Episode list, download, play status endpoints
│   │   ├── tag.go             # Tag CRUD endpoints
│   │   ├── search.go          # Search (iTunes, PodcastIndex, gPodder) endpoints
│   │   ├── settings.go        # Settings read/write endpoints
│   │   ├── backup.go          # Backup list/create endpoints
│   │   └── opml.go            # OPML import/export endpoints
│   ├── dto/                   # NEW — API response data transfer objects
│   │   ├── podcast.go
│   │   ├── episode.go
│   │   ├── tag.go
│   │   ├── settings.go
│   │   └── common.go          # Pagination envelope, error responses
│   └── router.go              # NEW — /api/v1 route registration
├── spa/                       # NEW — React SPA
│   ├── embed.go               # //go:embed dist/*
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   ├── router.tsx         # React Router config
│   │   ├── api/               # API client (fetch wrapper)
│   │   ├── components/        # Shared UI components
│   │   ├── pages/             # Route-level page components
│   │   ├── hooks/             # Custom React hooks
│   │   └── ws/                # WebSocket client
│   └── dist/                  # Build output (embedded in Go binary)
├── client/                    # Existing — SSR templates (removed in Phase H)
├── webassets/                 # Existing — vendored JS/CSS (removed in Phase H)
└── internal/sanitize/         # Existing
```

## Sources

- Martin Fowler, "Strangler Fig Pattern" — https://martinfowler.com/bliki/StranglerFigApplication.html (HIGH confidence)
- Go `//go:embed` documentation — https://pkg.go.dev/embed (HIGH confidence, Go 1.16+ standard library)
- Gin framework routing groups — https://gin-gonic.com/docs/examples/routing-group (HIGH confidence, already in project)
- Codebase analysis of `main.go`, `controllers/`, `service/`, `db/` (HIGH confidence, first-hand analysis)
- Existing route behavior verified by reading `controllers/podcast.go` return types (HIGH confidence)

---
*Architecture research: 2026-05-12*
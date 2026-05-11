# Technology Stack

**Project:** Podgrab — Self-hosted podcast manager modernization
**Researched:** 2026-05-12
**Context:** Brownfield Go app upgrading from Go 1.15/Gin 1.7/GORM 1.20 + Vue 2 SSR to Go 1.24+/Gin 1.12/GORM 1.26 + React SPA

## Recommended Stack

### Core Language & Runtime

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Go | 1.24.x | Backend language | Stable LTS (supported until Go 1.26+). Jumping from 1.15 → 1.24 gives range-over-int, enhanced routing in `net/http`, structured logging `log/slog`, `errors.Join`, generic slices/maps. 1.25 is current but 1.24 is more battle-tested and still fully supported. Don't go to 1.26 (released Feb 2026) yet — let it bake one more patch cycle. | HIGH |
| Node.js | 22.x LTS | Frontend build runtime | Required for Vite/React toolchain. 22.x is current LTS. | HIGH |

### Backend Framework

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Gin | v1.12.0 | HTTP framework | Latest stable. Minimum Go 1.21 per Gin 1.11+. Adds BSON protocol, protobuf content negotiation, better binding, HTTP/3 support. Direct upgrade from 1.7.2 — API is backward-compatible. | HIGH |
| Gin Router Groups | (built-in) | REST API route organization | Use `r.Group("/api/v1")` pattern for versioned REST endpoints alongside existing template routes. This is Gin's idiomatic API structure — no extra library needed. | HIGH |
| Gorilla WebSocket | v1.5.3 | WebSocket for player sync | Latest release. Project already uses Gorilla WebSocket — stay with it. v1.5.3 reverted some problematic v1.5.2 changes back to stable v1.5.0 baseline with bugfixes. | HIGH |

### ORM & Database

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| GORM | v1.26.0 | ORM | Latest stable in 1.25.x line. Adds Generics API (`gorm.G[Model](db)`) for type-safe queries in Go 1.18+. Traditional API still works — migrate incrementally. v1.31.1 is newest but v1.26 is proven and avoids genercis churn during migration. | HIGH |
| GORM SQLite Driver | v1.5.7 | SQLite dialect | Current latest. Compatible with GORM v1.25+. Requires CGO (via `mattn/go-sqlite3`). | HIGH |
| SQLite | 3.x (via CGO) | Database | Must stay SQLite — project constraint. GORM SQLite driver uses `mattn/go-sqlite3` which requires `CGO_ENABLED=1` and `gcc`. | HIGH |

### Backend Infrastructure Libraries

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| `go.uber.org/zap` | v1.28.0 | Structured logging | Replace `fmt.Println` with proper structured logging. Already a dependency (v1.16.0) — upgrade and actually *use* it this time. Consider `log/slog` (stdlib, Go 1.21+) for new code as alternative, but zap is already in the codebase and offers better performance. | HIGH |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT auth | Replace `dgrijalva/jwt-go` (abandoned, v3.2.0+incompatible). `golang-jwt/jwt` is the community-maintained successor. v5 is current and API-stable. | HIGH |
| `github.com/google/uuid` | v1.6.0 | UUID generation | Replace `satori/go.uuid` (abandoned, has known bugs). `google/uuid` is actively maintained and the standard replacement. | HIGH |
| `github.com/robfig/cron/v3` | v3.0.1 | Job scheduling | Replace `jasonlvhit/gocron` (stale, v0.0.1). `robfig/cron` is the Go standard for cron scheduling. v3 adds optional seconds field and `WithSeconds()` parser. | HIGH |
| `github.com/joho/godotenv` | v1.5.1 | `.env` file loading | Upgrade from v1.3.0. Minor but stays compatible. | HIGH |
| `golang.org/x/crypto` | latest | Bcrypt, crypto functions | Must update — security-critical. No version pin needed; use `go get -u` for latest. | HIGH |
| `github.com/microcosm-cc/bluemonday` | v1.0.26 | HTML sanitization | Upgrade from v1.0.15. Needed for RSS feed description sanitization. | MEDIUM |
| `github.com/antchfx/xmlquery` | latest | XML DOM querying | Keep — used for RSS feed image extraction. Upgrade to latest. | MEDIUM |
| `github.com/TheHippo/podcastindex` | v1.0.0 | PodcastIndex API client | Keep. Small library, no known newer alternative. Pin current version. | MEDIUM |

### Frontend Framework

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| React | 19.2.x | UI library | React 19 is current (released Dec 2024). Adds Actions, `use()`, improved Suspense, React Compiler support. React 19.2 is latest patch. Skip React 18 — 19 is stable and well-adopted now. | HIGH |
| React DOM | 19.2.x | DOM rendering | Paired with React 19. | HIGH |
| TypeScript | 6.0.x | Type safety | Current stable. Essential for maintainability of a React SPA. | HIGH |
| Vite | 8.0.x | Build tool, dev server | Current latest. Vite 8 released alongside `@vitejs/plugin-react` v6 (uses Oxc for React Refresh, no Babel). Instant HMR, fast builds. | HIGH |
| `@vitejs/plugin-react` | v6.0.x | Vite React plugin | Current. Uses Oxc transformer for fast refresh — no Babel dependency. Supports React Compiler opt-in. | HIGH |

### Frontend Ecosystem

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| React Router | v7.15.x | Client-side routing | v7 is current (remix merged). Works as a pure client-side router — no SSR needed. v7 unifies the Remix data loading patterns. Don't use v6 — v7 is the current line. | HIGH |
| TanStack Query | v5.100.x | Server state management | Handles API data fetching, caching, background refetching, optimistic updates. Eliminates most hand-rolled fetch state. Essential for a SPA that talks to a REST API. | HIGH |
| Zustand | v5.0.x | Client state management | Lightweight alternative to Redux. Perfect for client-only state (player state, UI preferences). Only 1KB. Don't use Redux — it's overkill for a self-hosted podcast manager. | HIGH |
| Tailwind CSS | v4.3.x | Utility-first CSS | v4 is a ground-up rewrite using Rust-based engine (Oxide), CSS-native config, no `tailwind.config.js`. Faster builds, simpler setup. Uses `@tailwindcss/vite` plugin for Vite integration. | HIGH |
| `@tailwindcss/vite` | v4.3.x | Tailwind Vite plugin | Official Vite plugin for Tailwind CSS v4. Required — replaces PostCSS-based approach from v3. | HIGH |
| shadcn/ui | v4.7.x | Component library | Copy-paste components (not npm deps) built on Radix UI + Tailwind. Customize by owning the code. Perfect for a self-hosted tool — don't need a full design system. | MEDIUM |
| Lucide React | v1.14.x | Icon library | Clean, consistent icons. Used by shadcn/ui. Replaces whatever icon approach the old UI used. | MEDIUM |
| Axios | v1.16.x | HTTP client | Already used in the Vue frontend. Keep it for the React frontend too — consistent API call patterns with interceptors for auth. Could use `fetch` directly, but Axios interceptors simplify auth token handling. | MEDIUM |
| Luxon | latest | Date/time formatting | Already used in Vue frontend. Keep for consistency — library handles podcast pub dates, duration formatting, etc. | MEDIUM |

### Development Tooling

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| ESLint | v10.3.x | Code linting | Current flat-config version. | HIGH |
| `typescript-eslint` | latest | TypeScript ESLint rules | Required for TypeScript-aware linting. Replaces `@typescript-eslint/eslint-plugin`. | HIGH |
| Docker | existing | Container build/runtime | Keep existing multi-stage Dockerfile pattern. Update base images (Go builder, Alpine runtime). | HIGH |
| Docker Compose | v2.x | Container orchestration | Keep existing pattern. | HIGH |

## What NOT to Use

| Rejected | Category | Why Not |
|----------|----------|--------|
| Next.js | React framework | Overkill. Podgrab is a Go backend + SPA. No SSR, no server components needed. Next.js adds complexity with no benefit when the backend is Go. |
| Remix | React framework | Same reason as Next.js. Remix's data loading patterns are useful but React Router v7 (which merged Remix) gives you the router without the server framework. |
| Redux / Redux Toolkit | State management | Overkill for a single-user podcast manager. Zustand covers client state; TanStack Query covers server state. Redux's boilerplate is not worth it here. |
| `satori/go.uuid` | UUID generation | Abandoned, has collision bugs. Use `google/uuid`. |
| `dgrijalva/jwt-go` | JWT auth | Abandoned, security issues. Use `golang-jwt/jwt/v5`. |
| `jasonlvhit/gocron` | Job scheduling | Stale (v0.0.1, unmaintained). Use `robfig/cron/v3`. |
| Vue 3 | Frontend framework | User chose React. Don't try to incrementally migrate Vue 2 → Vue 3 — that's a full rewrite anyway (Composition API, different reactivity). The decision is React. |
| Webpack | Build tool | Dead for new projects. Vite is faster, simpler, and the standard. |
| `nhooyr/websocket` | WebSocket library | Now `coder/websocket` v1.8.x. Fine library, but project already uses Gorilla WebSocket which is back under active maintenance. No reason to switch WebSocket libraries during a migration — that's unnecessary churn. |
| Prisma / Drizzle | ORM (frontend) | Frontend doesn't talk to DB — it talks to REST API. No ORM needed on the frontend side. |
| Material UI / Ant Design | Component library | Heavy, opinionated design systems that fight with Tailwind. shadcn/ui is copy-paste, customizable, and Tailwind-native. |
| `log/slog` (stdlib) | Structured logging | Available in Go 1.21+ and a valid choice for new projects. But zap is already in the codebase (v1.16.0) and offers better performance. Mixing slog + zap adds confusion. Standardize on zap during migration. Can reconsider slog in a future milestone. |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Backend framework | Gin 1.12 | Echo | Echo is fine but project is already on Gin. Migration within Gin is simpler than switching frameworks. Gin 1.12 has everything needed (API grouping, middleware, versioning). |
| Backend framework | Gin 1.12 | Chi | Chi is a lighter router. But it means rewriting all routes and losing Gin's binding/validation middleware. Not worth the churn. |
| Backend framework | Gin 1.12 | stdlib `net/http` | Go 1.22+ has improved routing. Could work for new routes but would require rewriting all existing routes. Too much churn for a migration milestone. |
| State management (client) | Zustand | Jotai | Jotai is atomic/model-based, Zustand is store-based. For a podcast manager with distinct state domains (player, subscriptions, UI), stores map better. |
| State management (server) | TanStack Query | SWR | TanStack Query has better devtools, mutation support, and cache management. SWR is simpler but less capable for podcast download status tracking. |
| CSS framework | Tailwind CSS v4 | CSS Modules | Tailwind offers faster development velocity for a small team. CSS Modules are more structured but verbose. For a self-hosted tool, velocity wins. |
| Component library | shadcn/ui | Headless UI + custom | shadcn/ui is Headless UI (Radix) + Tailwind + copy-paste. Same approach but faster to ship. |
| UUID generation | `google/uuid` | `github.com/lithammer/shortuuid` | `google/uuid` is the standard. Short UUID changes the format which could break existing DB records. |
| Cron scheduling | `robfig/cron/v3` | `github.com/reugn/go-quartz` | `robfig/cron` is simpler and the de facto standard. `go-quartz` adds persistence and clustering we don't need. |

## Go Dependency Upgrade Path

| Current | Target | Breaking Changes | Migration Effort |
|---------|--------|-------------------|-------------------|
| Go 1.15 → 1.24 | Go 1.24.x | Range-over-int syntax (opt-in), `io/ioutil` deprecated, new `any` alias | Medium — update deprecated calls, test thoroughly |
| Gin 1.7.2 → 1.12.0 | Gin 1.12.0 | Minimal API changes. `ShouldBind` improvements. Go 1.21 minimum. | Low — mostly drop-in |
| GORM 1.20.2 → 1.26.0 | GORM 1.26.0 | New generics API is opt-in. Traditional API mostly compatible. `gorm.Model` unchanged. | Low-Medium — test migrations, check callbacks |
| `dgrijalva/jwt-go` v3 → `golang-jwt/jwt/v5` | v5.3.1 | Different import path, API slightly changed (claims methods). | Medium — rewrite JWT usage |
| `satori/go.uuid` → `google/uuid` | v1.6.0 | Different import path, `uuid.New()` returns `[16]byte` not string. | Medium — update all model primary key generation |
| `jasonlvhit/gocron` → `robfig/cron/v3` | v3.0.1 | Completely different API. Cron syntax instead of chained scheduler. | Medium — rewrite scheduling setup |
| `go.uber.org/zap` v1.16.0 → v1.28.0 | v1.28.0 | Largely compatible. `SugaredLogger` still available. | Low — drop-in upgrade, use more |
| Gorilla WebSocket 1.4.2 → 1.5.3 | v1.5.3 | Minor API additions. `http.ResponseController` support. | Low — mostly drop-in |

## Frontend Package Installation

```bash
# Create React frontend project
npm create vite@latest frontend -- --template react-ts

# Core
cd frontend
npm install react@19 react-dom@19 react-router@7 @tanstack/react-query zustand

# HTTP & Utilities
npm install axios luxon

# Styling
npm install -D tailwindcss @tailwindcss/vite
npx shadcn@latest init

# Icons
npm install lucide-react

# Type definitions (dev)
npm install -D @types/react @types/react-dom typescript

# Linting (dev)
npm install -D eslint typescript-eslint
```

## Go Dependency Installation

```bash
# Upgrade core
go get go1.24    # or edit go.mod and run go mod tidy
go get github.com/gin-gonic/gin@v1.12.0
go get gorm.io/gorm@v1.26.0
go get gorm.io/driver/sqlite@v1.5.7

# Replace deprecated dependencies
go get github.com/golang-jwt/jwt/v5@v5.3.1
go get github.com/google/uuid@v1.6.0
go get github.com/robfig/cron/v3@v3.0.1

# Upgrade existing
go get go.uber.org/zap@v1.28.0
go get github.com/gorilla/websocket@v1.5.3
go get github.com/joho/godotenv@v1.5.1
go get golang.org/x/crypto  # latest
go get github.com/microcosm-cc/bluemonday  # latest
go get github.com/antchfx/xmlquery  # latest

# Remove deprecated
go mod edit -droprequire=github.com/dgrijalva/jwt-go
go mod edit -droprequire=github.com/satori/go.uuid
go mod edit -droprequire=github.com/jasonlvhit/gocron

go mod tidy
```

## Architecture Notes

### REST API Layer

Add `/api/v1/` route group to existing Gin router. Existing template routes stay unchanged during migration.

```
/api/v1/podcasts          → CRUD for podcast subscriptions
/api/v1/podcasts/:id/episodes → List episodes for podcast
/api/v1/search            → Search iTunes/PodcastIndex/gPodder
/api/v1/episodes/:id      → Episode details, download trigger
/api/v1/tags              → Tag management
/api/v1/settings          → App settings
/api/v1/auth/login        → Basic auth / JWT token
/ws/player                → WebSocket (existing, keep)
```

### Dual-UI Serving Pattern

During migration, both UIs coexist:
- `GET /` and old routes → Go template rendering (existing)
- `GET /app/*` → React SPA (served as static files from embedded or `/assets` volume)
- `GET /api/v1/*` → REST JSON endpoints (new)

Gin serves React SPA via `StaticFS` or embed the built frontend into the Go binary with `embed.FS`.

### Auth Strategy

Keep HTTP Basic Auth for template routes. Add JWT token auth for API routes:
- `POST /api/v1/auth/login` → validates basic auth credentials → returns JWT
- API middleware validates JWT on subsequent requests
- WebSocket auth stays as-is (query param or header)

## Sources

- **Go releases:** https://go.dev/doc/devel/release — Go 1.24.0 released 2025-02-11, Go 1.25.0 released 2025-08-12, Go 1.26.0 released 2026-02-10 (verified 2026-05-12)
- **Gin releases:** https://github.com/gin-gonic/gin/releases — v1.12.0 released 2026-02-28, minimum Go 1.24 (verified 2026-05-12)
- **GORM releases:** https://github.com/go-gorm/gorm/releases — v1.31.1 latest, v1.26.0 recommended target (verified 2026-05-12)
- **GORM SQLite driver:** Go proxy — v1.5.7 latest (verified 2026-05-12)
- **Gorilla WebSocket releases:** https://github.com/gorilla/websocket/releases — v1.5.3 latest (verified 2026-05-12)
- **React:** npm registry — v19.2.6 current (verified 2026-05-12)
- **Vite:** npm registry — v8.0.12 current (verified 2026-05-12)
- **Vite React plugin:** Context7 docs — v6.0.x uses Oxc transformer (verified 2026-05-12)
- **Tailwind CSS:** npm registry — v4.3.0 current with `@tailwindcss/vite` (verified 2026-05-12)
- **golang-jwt:** Go proxy — v5.3.1 latest (verified 2026-05-12)
- **google/uuid:** Go proxy — v1.6.0 latest (verified 2026-05-12)
- **robfig/cron:** Go proxy — v3.0.1 latest (verified 2026-05-12)
- **go.uber.org/zap:** Go proxy — v1.28.0 latest (verified 2026-05-12)
- **React Router:** npm registry — v7.15.0 (remix merged) (verified 2026-05-12)
- **TanStack Query:** npm registry — v5.0.13 (verified 2026-05-12)
- **Zustand:** npm registry — v5.0.13 (verified 2026-05-12)

---
*Stack research: 2026-05-12*
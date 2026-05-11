# Feature Landscape

**Domain:** Self-hosted podcast manager (download-first, single-user)
**Researched:** 2026-05-12
**Confidence:** HIGH (verified against Audiobookshelf docs, PodFetch docs, Podverse README, original Podgrab issues, and current codebase analysis)

## Competitive Landscape Summary

| App | Stars | Stack | Podcast-Only | Key Differentiator |
|-----|-------|-------|-------------|---------------------|
| **Audiobookshelf** | 8k+ | Node.js + SQLite | No (audiobooks + podcasts) | Multi-user, mobile apps, PWA, cross-device sync |
| **PodFetch** (Rust) | 476 | Rust + React + SQLite/Postgres | Yes | gPodder integration, S3 storage, mobile app |
| **Podverse** | 100 | React/Next.js | Yes (cloud-hosted) | Podcasting 2.0, chapters, clips, playlists, transcripts |
| **Podgrab** (original) | 2k | Go + Vue 2 + SQLite | Yes | Simplicity, download-first, NFO metadata |

Podgrab's niche is clear: **download-first, single-user, self-hosted Go app**. It competes with Audiobookshelf (heavy, audiobook-focused) and PodFetch-Rust (more active, Rust rewrite). The modernization must bring Podgrab's UI QoL up to parity without bloating it into Audiobookshelf.

---

## Table Stakes

Features users expect. Missing = product feels incomplete compared to Audiobookshelf/PodFetch. Every modern self-hosted podcast manager has these.

| Feature | Why Expected | Complexity | Podgrab Status | Notes |
|---------|--------------|------------|----------------|-------|
| **Responsive design** | Users access from phones/tablets, not just desktop | Med | ❌ Vue 2 SSR templates are not responsive | Mobile viewport is non-negotiable in 2026; Audiobookshelf and PodFetch are fully responsive |
| **Playback speed control** | 1x/1.25x/1.5x/2x is standard in every podcast app since 2018 | Low | ❌ Not in current player | HTML5 `playbackRate` property — trivial to add, embarrassing to omit |
| **Episode play progress tracking** | Users expect to resume where they left off | Med | ⚠️ WebSocket sync exists across tabs but no persistent progress | Must persist to DB; Audiobookshelf does per-user per-episode progress |
| **Played/unplayed episode indicator** | Visual distinction between new/listened/completed episodes | Low | ❌ No concept of "played" state | Podgrab original roadmap listed this as planned; every competitor has it |
| **Episode search/filter** | 100+ episodes per podcast; need to find specific ones | Med | ❌ No search/filter in current UI | Filter by title, date range, download status, played status |
| **Episode sorting** | Users want chronological, newest-first, oldest-first | Low | ⚠️ Limited — current UI doesn't expose sort controls | Must support: newest first, oldest first, alphabetical |
| **Download status indicators** | Users need to know: queued / downloading / complete / failed | Low | ⚠️ Basic — "downloaded" flag exists, but no progress or queue visibility | Need: not downloaded, queued, downloading (with %), complete, error |
| **Dark mode** | Every modern self-hosted app has it; Podgrab original had it | Low | ❌ Was in original Vue 2 UI, lost in migration? | Use CSS `prefers-color-scheme` + manual toggle; store preference |
| **Keyboard shortcuts** | Power users expect space (play/pause), arrows (seek), etc. | Low | ❌ None | Volume, seek ±15s/±30s, play/pause, next episode, speed |
| **Persistent audio player** | Player must survive navigation between pages | Med | ⚠️ Current player is page-bound; only sync works across tabs | Sticky bottom bar (like Spotify Web / Audiobookshelf) — player stays visible while browsing |
| **Episode description/notes** | Show RSS show notes inline on episode detail | Low | ❌ Episode detail doesn't render show notes | RSS feed has `<description>` with HTML — need sanitized rendering |

## Differentiators

Features that set Podgrab apart or bring it above the competition. Not universally expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Per-podcast auto-download rules** | Users want: auto-download latest 3 only, auto-download all, or manual-only per podcast | Med | Audiobookshelf has per-podcast auto-download schedule. PodFetch has global polling interval. Podgrab currently has global auto-download only — per-podcast control is a differentiator |
| **Disk space management** | Auto-cleanup: delete played episodes after X days, keep max N episodes per podcast | Med | No competitor does this well — Podgrab's download-first nature makes storage management critical. NFO metadata preserves info even after file deletion |
| **RSS feed re-scan on demand** | Button to force-refresh a podcast's feed immediately | Low | Competitors have cron-only refresh. Per-podcast manual refresh is QoL that power users expect |
| **Batch episode actions** | Select multiple episodes → mark played, delete, download | Med | Audiobookshelf has episode bulk download; no one has bulk mark-played or bulk delete well |
| **Subscription health / feed errors** | Show when RSS feeds are broken, 403'd, or haven't updated in N days | Med | Original Podgrab issues (#225, #313) are about feed failures with no user feedback. Surface errors clearly |
| **Episode timeline / activity feed** | "What's new" view across all subscriptions, newest-first | Med | Audiobookshelf and PodFetch have a "latest episodes" dashboard; Podgrab needs an inbox-style view |
| **gPodder sync (bi-directional)** | Sync subscriptions and episode state with gPodder-compatible mobile apps (AntennaPod, etc.) | High | PodFetch has this. Podgrab already has gPodder search — extending to full sync (subscribe on phone → appears in Podgrab) would be a differentiator for the existing Go gPodder code |
| **PWA support** | Install to home screen, offline-ish with service worker for cached UI | Med | Audiobookshelf and PodFetch have PWA. Not hard with Vite + React; big UX win on mobile |
| **Podcast 2.0 features** | Chapters, transcripts, value4value — Podverse leads here | High | Forward-looking; chapters are most impactful (skip intro/outro, see segments). Start with chapters from RSS `<podcast:chapters>` |
| **Streaming playback (without full download)** | Play episode from remote URL without waiting for download | Med | Current Podgrab is download-first only; streaming is explicitly out-of-scope in PROJECT.md — but a "stream while downloading" hybrid is high value and not that complex since files are being downloaded to disk anyway |

## Anti-Features

Features to explicitly NOT build. These either contradict Podgrab's identity or are resource traps.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Multi-user / RBAC** | Podgrab is single-user. Adding multi-user means auth overhaul, per-user state, permission system — Audiobookshelf's entire architecture revolves around this | Keep HTTP Basic Auth; if someone needs multi-user, they should use Audiobookshelf |
| **Audiobook management** | Audiobookshelf exists for this. Mixing audiobooks into Podgrab doubles the domain complexity (chapter metadata, author folders, m4b merging, ebook support) | Stay podcast-only. If someone asks, point to Audiobookshelf |
| **Social features** (comments, sharing, likes, follower counts) | Not core to download-and-organize. Podverse does social; that's their differentiator | Focus on the download/organize/listen loop |
| **Native mobile app** | Massive investment (React Native or separate Swift/Kotlin codebases). Audiobookshelf spent years on their mobile app | PWA is 80% of the benefit at 10% of the cost |
| **Music / general media library** | Jellyfin/Plex exist. Don't scope-creep into being a media server | Podcast download manager, not media server |
| **AI features** (auto-transcription, summarization, recommendation engine) | Cool but orthogonal to core value prop; requires GPU/cloud API costs; maintenance burden | Could be a future plugin system, but not core |
| **Podcast monetization / value4value / Bitcoin** | Podverse's niche. Implementation requires wallet integrations, payment APIs | Not Podgrab's lane |
| **HLS streaming support** | Episode #1402 on PodFetch. HLS is rarely used by podcasts; adds significant player complexity | Support standard MP3/AAC/M4A progressive download; if HLS needed, that's a new major feature |

## Feature Dependencies

```
Episode play progress tracking ──→ Played/unplayed indicator
         │                              │
         ▼                              ▼
   Persistent audio player        Episode search/filter
         │                         (filter by played status)
         │
         ▼
   Playback speed control
   Keyboard shortcuts

Per-podcast auto-download ──→ Disk space management
         │                        (auto-cleanup uses per-podcast rules)
         ▼
   RSS feed re-scan on demand

Batch episode actions ──→ Episode search/filter
         │              (need filter to select batch)
         ▼
   Subscription health
   (batch mark downloaded-errored feeds)

Subscription health ──→ Episode timeline / activity feed
         │              (show feed errors in timeline)
         ▼
   gPodder sync
   (sync error states across devices)
```

## MVP Recommendation

**Priority 1 — Parity (must ship with React migration):**
1. Responsive design (without this, the migration is a downgrade for mobile users)
2. Persistent audio player with playback speed control
3. Episode play progress tracking (persist to DB)
4. Played/unplayed episode indicators
5. Dark mode
6. Episode sorting (newest/oldest/alpha)
7. Download status indicators (queued/downloading/complete/error)

**Priority 2 — QoL (ship ASAP after React migration stabilizes):**
1. Episode search/filter
2. Keyboard shortcuts
3. Episode description/notes rendering
4. Per-podcast auto-download rules
5. RSS feed re-scan on demand
6. Episode timeline / "what's new" dashboard

**Priority 3 — Differentiators (build when core is solid):**
1. Batch episode actions
2. Disk space management / auto-cleanup
3. PWA support
4. Subscription health / feed error visibility
5. gPodder bi-directional sync

**Defer indefinitely:** Multi-user, audiobooks, social, native mobile, AI, value4value, HLS

## Gap Analysis: What Podgrab Has That Competitors Don't

| Feature | Podgrab Only | Notes |
|---------|-------------|-------|
| NFO metadata generation | Yes | Kodi/Plex-friendly episode metadata — unique to Podgrab's download-first philosophy |
| Go backend | Yes (vs Node.js/Rust) | Lower memory footprint, single binary, easier self-hosting for Go-familiar users |
| WebSocket player sync across tabs | Yes (basic) | Uniquely real-time in the self-hosted space; needs persistence but the sync concept is good |

These should be **preserved and improved** during modernization, not dropped.

## Modern UI Patterns From Competitors

### Audiobookshelf Patterns Worth Emulating
- **Sticky bottom player bar**: Player bar stays visible during all navigation; progress bar is scrubbbable
- **Library → Podcast → Episode hierarchy**: Three-level drill-down with breadcrumb navigation
- **Per-episode progress bars**: Visual progress (0-100%) overlaid on episode rows in list view
- **Auto-download schedule per podcast**: Crontab-style scheduling for each subscription
- **Episode queue / playlist**: Add episodes to a listening queue separate from subscription list

### PodFetch Patterns Worth Emulating
- **Dashboard view**: Shows newest episodes across all subscriptions at a glance
- **gPodder compatibility**: Sync with AntennaPod, etc. for mobile + server parity
- **Podcast-specific settings**: Auto-download toggle, episode limit per podcast

### Podverse Patterns (Forward-Looking)
- **Chapter markers**: Clickable chapter list from RSS `<podcast:chapters>` or chapters JSON
- **Clip creation**: Share a timestamped segment of an episode (social, but useful for personal bookmarks)
- **Transcript display**: Show transcript alongside playback (if available in feed)

## Sources

- **Audiobookshelf**: Official docs (audiobookshelf.org/docs), Context7 API docs (HIGH confidence)
- **PodFetch**: GitHub README, docs site (podfetch-docs.samtv.fyi), GitHub issues (HIGH confidence)
- **Podverse**: GitHub README (podverse/podverse-web) (MEDIUM confidence — cloud-hosted model differs from self-hosted)
- **Podgrab (original)**: GitHub README, issue tracker (akhilrex/podgrab/issues) — especially #23 (ID3 tags), #154 (offline sync), #225/#313 (feed errors) (HIGH confidence)
- **Current codebase**: .planning/codebase/ARCHITECTURE.md analysis (HIGH confidence)
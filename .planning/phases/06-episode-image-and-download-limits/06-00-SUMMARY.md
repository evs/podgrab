---
phase: 06
plan: 00
subsystem: episode-image-endpoint
tags: [image, fallback, endpoint, episodes]
dependency_graph:
  requires: []
  provides: [episode-image-fallback]
  affects: [controllers/podcast.go]
tech_stack:
  added: []
  patterns: [fallback-chain, cascade-pattern]
key_files:
  created: []
  modified:
    - controllers/podcast.go
decisions:
  - Used existing Preload(clause.Associations) from GetPodcastItemById instead of adding a secondary query — avoids N+1
  - Redirect to /webassets/blank.png as final fallback instead of returning 404 — consistent with existing onImageError JS handler
metrics:
  duration: 5min
  completed: 2026-05-14
---

# Phase 6 Plan 00: Episode Image Endpoint Fallback Summary

Added parent podcast image fallback to the episode image endpoint, fixing broken image icons in episode lists when episodes lack individual artwork.

## What Changed

Modified `GetPodcastItemImageById` in `controllers/podcast.go` to implement a four-tier fallback chain:

1. **Local image file** — If the episode has a `LocalImage` path and the file exists on disk, serve it directly
2. **Episode-level image URL** — If the episode has an `Image` URL, redirect to it (original behavior)
3. **Parent podcast cover image** — If the episode has no image, redirect to `podcastItem.Podcast.Image` (the parent podcast's cover art)
4. **Blank placeholder** — If no image is available at any level, redirect to `/webassets/blank.png` (same fallback as the existing `onImageError` JS handler)

## Why This Works

- `db.GetPodcastItemById` already uses `DB.Preload(clause.Associations)` which eagerly loads the `Podcast` association
- `podcastItem.Podcast.Image` is therefore available without any additional query — zero performance overhead
- The fallback chain preserves the existing behavior for episodes that already have images

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1-5 | Episode image fallback with parent podcast cover | b808f98 | controllers/podcast.go |

## Verification

- `go build ./...` passes
- `go vet ./...` passes
- The `GetPodcastItemById` function already preloads associations, so `podcastItem.Podcast.Image` is populated

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Threat Flags

None — no new endpoints or trust boundaries introduced.

## Self-Check: PASSED

- FOUND: controllers/podcast.go
- FOUND: commit b808f98
- FOUND: 06-00-SUMMARY.md
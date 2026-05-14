---
phase: 06
plan: 01
subsystem: data-model
tags: [gorm, model, settings, ui, max-episodes]
dependency_graph:
  requires: []
  provides: [MaxEpisodes-column, settings-ui]
  affects: [db/podcast.go, service/podcastService.go, controllers/pages.go, controllers/podcast.go, client/settings.html]
tech_stack:
  added: [gorm-column-default]
  patterns: [model-field-extension, settings-pass-through]
key_files:
  created: []
  modified:
    - db/podcast.go
    - service/podcastService.go
    - controllers/pages.go
    - controllers/podcast.go
    - client/settings.html
decisions:
  - "MaxEpisodes defaults to 0 meaning unlimited — preserves backward compatibility"
  - "Added to both Setting (global default) and Podcast (per-podcast override) for future flexibility"
  - "Numeric input with min=0 on settings page — 0 means keep all episodes"
metrics:
  duration: 10min
  completed: "2026-05-14"
---

# Phase 06 Plan 01: MaxEpisodes Model & Settings UI Summary

Add `MaxEpisodes` column to both `Setting` (global default) and `Podcast` (per-podcast override) models for future episode count limits, with a numeric input on the settings page.

## What Was Done

1. **db/podcast.go** — Added `MaxEpisodes int gorm:"default:0"` to both `Podcast` struct (line 35) and `Setting` struct (line 92). Default 0 means "keep all episodes" for backward compatibility.

2. **service/podcastService.go** — Added `maxEpisodes int` parameter to `UpdateSettings()` function signature and assigned `setting.MaxEpisodes = maxEpisodes` before persisting.

3. **controllers/pages.go** — Added `MaxEpisodes int` field to `SettingModel` struct with `form/json/query` tags for request binding.

4. **controllers/podcast.go** — Passed `model.MaxEpisodes` as the last argument to `service.UpdateSettings()` call in the `UpdateSetting` handler.

5. **client/settings.html** — Added:
   - Numeric input field with label "Maximum episodes to keep per podcast (0 = keep all)"
   - Vue.js `v-model.number` binding (`maxEpisodes`)
   - axios POST payload inclusion (`maxEpisodes: self.maxEpisodes`)
   - Go template data binding (`maxEpisodes: {{ .setting.MaxEpisodes }}`)

## Verification

- `go build ./...` compiles successfully
- `go test ./...` — all 4 packages pass (podgrab, controllers, db, service)
- GORM auto-migration will add the `MaxEpisodes` column on next app start

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Parallel executor merge conflict resolution**
- **Found during:** Task execution
- **Issue:** Parallel 06-02 executor modified service/podcastService.go and db/podcast.go, overwriting some of my edits
- **Fix:** Re-applied all changes after the parallel commit, rebuilt, re-tested, and committed as a single atomic commit
- **Files modified:** db/podcast.go, service/podcastService.go, controllers/pages.go, controllers/podcast.go, client/settings.html
- **Commit:** 1ba5ec1

## Known Stubs

None — the MaxEpisodes column is created and persisted but no enforcement logic is implemented yet. That is deliberate: plan 06-02 adds the `EnforcePodcastEpisodeLimit` business logic.

## Self-Check: PASSED

- [x] db/podcast.go — FOUND
- [x] service/podcastService.go — FOUND
- [x] controllers/pages.go — FOUND
- [x] controllers/podcast.go — FOUND
- [x] client/settings.html — FOUND
- [x] Commit 1ba5ec1 — FOUND
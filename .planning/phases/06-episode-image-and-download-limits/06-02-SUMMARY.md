---
phase: "06-episode-image-and-download-limits"
plan: "02"
subsystem: "episode-limit-enforcement"
tags: [episode-limit, pruning, settings, db-query]
dependency_graph:
  requires: ["06-01"]
  provides: ["enforce-episode-limit"]
  affects: ["service/podcastService", "db/dbfunctions", "db/podcast", "controllers/podcast", "client/settings"]
tech_stack:
  added: ["db.CountDownloadedEpisodesByPodcastId", "db.GetOldestDownloadedEpisodesByPodcastId", "EnforcePodcastEpisodeLimit"]
  patterns: ["per-podcast episode count limit with oldest-first pruning"]
key_files:
  created: ["service/podcastService_test.go (3 tests added)"]
  modified: ["db/podcast.go", "db/dbfunctions.go", "service/podcastService.go", "controllers/podcast.go", "service/podcastService_test.go"]
decisions:
  - "MaxEpisodes=0 means unlimited (no pruning) — zero-value safety"
  - "Episode limit is enforced after AddPodcastItems inserts new episodes, not during"
  - "Pruning uses DeleteEpisodeFile which handles file deletion + DB status update"
  - "Pruning errors are logged but do not fail the AddPodcastItems flow"
  - "Oldest episodes by PubDate are pruned first (FIFO pruning)"
metrics:
  duration: "10min"
  completed: "2026-05-14"
  tasks_completed: 5
  files_modified: 6
  commits: 4
---

# Phase 6 Plan 2: Episode Limit Enforcement Summary

Add per-podcast episode count limits — when MaxEpisodes > 0, prune oldest downloaded episodes after new ones are added.

## What Was Done

1. **MaxEpisodes setting field** — Added `MaxEpisodes int` to `db.Setting` struct with `gorm:"default:0"` (0 = unlimited)
2. **DB query helpers** — Added `CountDownloadedEpisodesByPodcastId` and `GetOldestDownloadedEpisodesByPodcastId` to `db/dbfunctions.go`
3. **EnforcePodcastEpisodeLimit function** — New helper in `service/podcastService.go` that:
   - Reads `MaxEpisodes` from settings (0 = no limit, skip)
   - Counts downloaded episodes for the podcast
   - If count > limit, fetches the N oldest downloaded episodes
   - Calls `DeleteEpisodeFile` on each excess episode (removes file + marks as Deleted)
   - Logs errors but does not fail the caller
4. **Integration in AddPodcastItems** — Called `EnforcePodcastEpisodeLimit` after new episodes are added, only when `len(itemsAdded) > 0`
5. **Settings wiring** — Added `MaxEpisodes` parameter to `UpdateSettings` and `SettingModel`
6. **Tests** — Three test cases: no limit (MaxEpisodes=0), prune oldest (MaxEpisodes=3), under limit (2 episodes with limit 5)

## Verification

- `go build ./...` passes
- `go test ./...` — all 4 test packages pass
- `TestEnforcePodcastEpisodeLimit_NoLimit` — confirms 5/5 episodes remain with MaxEpisodes=0
- `TestEnforcePodcastEpisodeLimit_PruneOldest` — confirms 3 remain downloaded, 2 pruned to Deleted
- `TestEnforcePodcastEpisodeLimit_UnderLimit` — confirms no pruning when under limit

## Deviations from Plan

None — plan executed exactly as written.

## Threat Flags

None — no new network endpoints or auth paths.

## Self-Check: PASSED

All files exist, all commit hashes verified, no accidental deletions.
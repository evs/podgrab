# Plan 02-04 Summary — Service and DB Layer Tests

**Status:** Complete
**Wave:** 1
**Requirements:** TEST-02, TEST-03

## What Changed
Added substantive CRUD test coverage for the DB and service layers.

### Files Modified
- `db/db_test.go` — Added 4 new test functions:
  - `TestGetPodcastByURL` — Create by URL, lookup by URL, verify not-found returns error
  - `TestGetAllPodcasts` — Create 3 podcasts, verify all are present in results
  - `TestPodcastItemsCRUD` — Create podcast + 2 items, verify 2 items found, delete 1, verify 1 remains
  - `TestBasicDBSetup` was already present from Wave 0

- `service/podcastService_test.go` — Added 3 new test functions:
  - `TestGetPodcastById_NotFound` — Verify nonexistent ID returns non-nil zero-value Podcast
  - `TestGetAllPodcasts_Service` — Create 2 podcasts via DB layer, verify service function retrieves both
  - `TestDeletePodcastEpisodes_ClearsItems` — Create podcast + 2 items, call `DeletePodcastEpisodes`, verify items have `DownloadStatus == Deleted`

### Key Adjustment
`TestGetAllPodcasts` uses presence-checking (map lookup) rather than strict count assertion because the `:memory:?cache=shared` DB is shared across tests.

### Verification
- `go test ./db/` exits 0 (5 tests)
- `go test ./service/` exits 0 (5 tests)
- `go test ./...` exits 0

## Commits
- `test(02-04): DB + Service layer CRUD tests`

# Plan 02-01 Summary — Test Harness Setup

**Status:** Complete
**Wave:** 0
**Requirements:** TEST-01

## What Changed
Created three test files establishing a working Go test harness for the podgrab project (zero tests existed before):

### Files Created
- `db/db_test.go` — DB layer test harness with `setupTestDB(t)` helper opening SQLite `:memory:?cache=shared`, plus `TestBasicDBSetup` verifying create-read-delete lifecycle
- `service/podcastService_test.go` — Service layer harness with `setupServiceTestDB(t)`, plus `TestParseOpml` (no DB needed) and `TestServiceTestDBSetup` verifying create-fetch-delete through service API
- `main_test.go` — Template function test with `TestRemoveStartingSlash`; initially asserts buggy behavior, updated in Plan 02-03

### Key Design Decision
`funcMap` in `main.go` was changed from local (`funcMap :=`) to package-level (`var funcMap =`) so `main_test.go` can reference it for direct testing of template functions.

## Verification
- `go test ./db/` exits 0
- `go test ./service/` exits 0
- `go test .` exits 0 (root package with main_test.go)
- `go test ./...` exits 0
- Zero external test dependencies — stdlib `testing` only

## Commits
- `test(02-01): Wave 0 — test harness setup with SQLite :memory:`

# Phase 4 Research: Error Handling Modernization

**Date:** 2026-05-13
**Phase:** 04-error-handling-modernization

## Findings

### ERR-01: fmt.Println / log.Println → structured zap logging

Scope: 25 occurrences across 6 files.

**service/podcastService.go** (14 occurrences):
- Lines 142, 244, 459, 528, 536, 589, 614, 635, 721, 733: `fmt.Println(err.Error())` or plain `fmt.Println(...)` used for errors and diagnostics.
- Lines 410, 560, 581, 603: Commented-out `fmt.Println` (pre-existing, leave as-is).
- Logger (`zap.SugaredLogger`) already initialized at top of file but used in only ~3 places.

**controllers/podcast.go** (5+ occurrences):
- Lines 92, 171, 184, 196, 220, 544, 632: `fmt.Println(err)` or `fmt.Println(err.Error())` after service calls.
- Many errors printed but NOT returned to client via `c.JSON(...)` — ERR-02 overlap.

**controllers/pages.go**:
- Lines 174, 387: `fmt.Println(err.Error())` and commented debug.

**controllers/websockets.go**:
- Lines 40, 131: `fmt.Println(...)` for WS errors.

**service/itunesService.go**:
- Lines 45, 53: `fmt.Println` for warnings and API errors.

**service/naturaltime.go**:
- Line 74: commented `fmt.Println`.

**main.go**:
- `log.Println(dbPath)` — pre-existing.

### ERR-04: checkError panic pattern

**service/fileService.go**, lines 404-408:
```go
func checkError(err error) {
    if err != nil {
        panic(err)
    }
}
```

Used once at line 386 in `getFileName()`:
```go
fileUrl, err := url.Parse(link)
checkError(err)
```

### ERR-02 / ERR-03: Error propagation and HTTP status codes

**controllers/podcast.go** pattern audit:
- **Silent errors** (only `fmt.Println(err)`, no JSON response): lines 92, 171, 184, 220. Client receives 200 OK with wrong/empty body.
- **Correct handler** (returns 400 + message): lines 103, 116, 395, 557, 569, 544, 632.
- **Mixed**: line 196 (`fmt.Println + return 400`).

**controllers/pages.go**:
- Lines 103, 106, 109, 229, 306, 342, 352, 358, 364, 383, 388: returns 400 on some errors, but some paths silently ignore errors.

**Notable problematic patterns:**
```go
// controllers/podcast.go:92
data, err := service.GetPodcastById(searchByIdQuery.Id)
if v, ok := err.(*model.PodcastAlreadyExistsError); ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Podcast already exists"})
    return err
}
c.JSON(200, data)
```
Wait, this pattern seems wrong — it's mixing `GetPodcastById` with `PodcastAlreadyExistsError`.

```go
// controllers/podcast.go:171
podcast, err := service.GetPodcastById(searchByIdQuery.Id)
fmt.Println(err)
c.JSON(200, podcast)  // returns 200 even on error
```

```go
// controllers/podcast.go:183
err := service.SetAllEpisodesToDownload(searchByIdQuery.Id)
fmt.Println(err)
c.JSON(200, gin.H{})
```

### Recommendations

1. Replace all `fmt.Println(err)` in service layer with `Logger.Errorw(field, err)`. Use the existing `Logger` instance.
2. In controllers, replace `fmt.Println(err)` with `c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})` or propagate the correct 4xx/5xx code.
3. Remove `checkError()` entirely; replace its single call site with normal error return. `getFileName` should return `(string, error)`.
4. For errors in HTTP handlers: if error is user-invalid (e.g., bad URL), return 400; if it's a server/internal error, return 500.

### Risks

- High file count touched (service/podcastService.go, service/itunesService.go, service/fileService.go, controllers/podcast.go, controllers/pages.go, controllers/websockets.go).
- Changing `getFileName` signature requires updating all 5+ call sites.
- Many `fmt.Println` are in non-error diagnostic paths (e.g., "Processing episodes: N"); these should become `Logger.Info` not `Logger.Error`.

## Test strategy

- Existing `go test` suite must continue to pass.
- No new tests needed — this is cleanup, not behavior change — but a verification grep for remaining `fmt.Println(err)` and `panic(` after changes is the acceptance gate.

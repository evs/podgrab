---
phase: 3
slug: correctness-concurrency-fixes
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-13
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (built-in) |
| **Config file** | none — Wave 0 installs in Phase 2 |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -race -v ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -race -v ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01 | 1 | BUG-04 | — | Download goroutines bounded by configurable limit | unit | `go test -v ./service/... -run TestDownloadConcurrency` | ❌ W0 | ⬜ pending |
| 3-01-02 | 01 | 1 | BUG-04 | — | Concurrency limit respects setting value (1, 3, 5) | unit | `go test -v ./service/... -run TestDownloadConcurrency` | ❌ W0 | ⬜ pending |
| 3-02-01 | 02 | 2 | BUG-05 | — | RSS dates parse for RFC1123, RFC3339, ISO 8601 | unit | `go test -v ./service/... -run TestParseRSSDate` | ❌ W0 | ⬜ pending |
| 3-02-02 | 02 | 2 | BUG-05 | — | Invalid dates return error (not silently zero) | unit | `go test -v ./service/... -run TestParseRSSDate` | ❌ W0 | ⬜ pending |
| 3-03-01 | 03 | 2 | BUG-06 | — | DB init error propagated (not nil DB) | unit | `go test -v ./db/... -run TestDBInit` | ❌ W0 | ⬜ pending |
| 3-03-02 | 03 | 2 | BUG-06 | — | DB ping succeeds before app continues | unit | `go test -v ./db/... -run TestDBInit` | ❌ W0 | ⬜ pending |
| 3-04-01 | 04 | 3 | BUG-07 | — | WebSocket map writes protected by mutex | unit/race | `go test -race -v ./controllers/... -run TestWebSocketMaps` | ❌ W0 | ⬜ pending |
| 3-04-02 | 04 | 3 | BUG-07 | — | No data races under concurrent load | race | `go test -race -v ./controllers/...` | ❌ W0 | ⬜ pending |
| 3-05-01 | 05 | 3 | TEST-04 | — | Download concurrency test cases pass | unit | `go test -v ./service/... -run TestDownloadConcurrency` | ❌ W0 | ⬜ pending |
| 3-05-02 | 05 | 3 | TEST-05 | — | Date parsing test cases pass | unit | `go test -v ./service/... -run TestParseRSSDate` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `service/episodeService_test.go` — stubs for download concurrency and date parsing tests
- [ ] `db/db_test.go` — DB init failure/success test cases
- [ ] `controllers/websocket_test.go` — WebSocket map race condition test cases
- [ ] `go test -race` enabled in default test workflow

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Download ordering across large podcast feed | BUG-04 | Requires real HTTP and filesystem I/O | Subscribe to a large podcast (>100 episodes), trigger download, observe max N concurrent downloads in logs |
| RSS date parsing on real-world feeds | BUG-05 | Variability in publisher date formats | Import 5 podcasts with different publishers, verify episode dates are non-zero in UI |

*All other behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

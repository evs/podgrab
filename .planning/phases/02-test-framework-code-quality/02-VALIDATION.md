---
phase: 2
slug: test-framework-code-quality
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-12
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5-10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 0 | TEST-01 | — | N/A | unit | `go test ./db/...` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 0 | TEST-01 | — | N/A | unit | `go test ./service/...` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02 | 1 | BUG-01 | — | N/A | static | `grep -c "DeletePodcasDeleteOnlyPodcastt" controllers/podcast.go` | ✅ | ⬜ pending |
| 02-03-01 | 03 | 1 | BUG-02 | — | N/A | unit | `go test -run TestRemoveStartingSlash` | ❌ W0 | ⬜ pending |
| 02-03-02 | 03 | 1 | BUG-03 | — | N/A | static | `grep -c 'fmt.Println(raw)' main.go` | ✅ | ⬜ pending |
| 02-04-01 | 04 | 1 | TEST-02 | — | N/A | unit | `go test ./db/...` | ❌ W0 | ⬜ pending |
| 02-04-02 | 04 | 1 | TEST-03 | — | N/A | unit | `go test ./service/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `db/db_test.go` — test stubs for DB CRUD
- [ ] `service/podcastService_test.go` — test stubs for service layer
- [ ] `main_test.go` — test for `removeStartingSlash`
- [ ] `go test ./...` compiles and passes

*Wave 0 installs the test harness so Wave 1+ tasks have something to verify against.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Template function behavior in HTML rendering | BUG-02 | Requires rendered HTML output | Build app, check URL paths in templates |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

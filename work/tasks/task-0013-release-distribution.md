---
id: task-0013
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0012]
doc_ref: technical/distribution/README.md
origin: rule
priority: low
depends_on: [task-0001]
milestone: null
created: 2026-09-03T22:30:25Z
queued: 2026-09-04T05:05:03Z
completed: null
merged: null
provenance: []
---

# Distribute the binary through Homebrew, releases and go install

**References:** [technical/distribution/README.md](../../docs/technical/distribution/README.md) · [spec-0012](../specs/spec-0012-release-distribution.md)

Ship the binary three ways: Homebrew tap, prebuilt binaries on tagged
GitHub releases, `go install`. GoReleaser builds the binaries and
updates the tap on each tagged release; the cut stays with
`scripts/release.sh`, and each release names the WritRun tag it
targets.

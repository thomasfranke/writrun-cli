---
id: task-0016
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0015]
doc_ref: technical/engineering/boundaries.md
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-05T12:28:30Z
queued: 2026-09-05T12:46:23Z
completed: null
merged: null
provenance: []
---

# Put the filesystem behind a port, as the boundaries rule requires

**References:** [technical/engineering/boundaries.md](../../docs/technical/engineering/boundaries.md) · [spec-0015](../specs/spec-0015-the-filesystem-is.md)

`boundaries.md` puts everything leaving the process behind a small
interface with a fake beside it. Script execution, `gh` and the terminal
are; the filesystem is not, in about forty call sites across the three
command packages.

It costs the suite what it costs the design. Coverage over `internal/`
reached 98.0% and stopped there: 18 of the 22 statements no fixture can
reach are error returns from those direct calls, and the tests that do
reach the reachable ones work by making a file read-only — shaping the
machine because there is no port to shape.

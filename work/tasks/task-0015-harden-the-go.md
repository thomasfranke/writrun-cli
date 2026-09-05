---
id: task-0015
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0014]
doc_ref: technical/testing/ci.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-05T11:18:19Z
queued: 2026-09-05T11:48:55Z
completed: 2026-09-05T16:46:09Z
merged: null
provenance: []
---

# Harden the Go pipeline and gate pull requests on coverage

**References:** [technical/testing/ci.md](../../docs/technical/testing/ci.md) · [spec-0014](../specs/spec-0014-every-pull-request.md)

A pull request may not be approved below 90% coverage, and the pipeline
should catch the classes of fault a Go project can catch cheaply before
a person reads the diff.

The gate that exists today is a single global percentage, and a global
percentage is exactly what a whole uncovered package hides: while
implementing task-0003 and task-0005, `internal/command/updatecmd` and
`internal/command/uninstallcmd` sat at 0% and the total stayed above
the 85% floor, so CI was green until the two packages were large enough
to drag the average down. What the pipeline reports has to be what a
reviewer would otherwise have to check by hand.

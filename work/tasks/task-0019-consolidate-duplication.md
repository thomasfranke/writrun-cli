---
id: task-0019
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0018]
doc_ref: technical/engineering/boundaries.md
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-05T19:11:12Z
queued: 2026-09-06T00:22:35Z
completed: 2026-09-06T00:46:13Z
merged: null
provenance: []
---

# Consolidate what the parallel round duplicated

**References:** [technical/engineering/boundaries.md](../../docs/technical/engineering/boundaries.md) · [spec-0018](../specs/spec-0018-consolidate-duplication.md)

Five tasks ran in parallel over the same packages and each took a
duplication rather than a coupling it could not pay for: three commands
now read `.writrun/VERSION`, two compare tags with implementations that
disagree on what they compute, and the stage-0 and stage-1 checks are
written twice with three behavioural differences between the copies.

All five have landed, so the reason is spent. Decide per duplication
whether it is extracted or kept — keeping is a valid answer when it is
cheaper than the coupling, as long as the reason is written where the
next reader finds it.

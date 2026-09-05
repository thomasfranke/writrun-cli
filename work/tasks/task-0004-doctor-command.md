---
id: task-0004
status: done
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0004]
doc_ref: product/adoption/doctor.md
origin: rule
priority: medium
depends_on: [task-0001]
milestone: null
created: 2026-09-03T22:30:18Z
queued: 2026-09-04T05:05:03Z
completed: 2026-09-05T18:22:30Z
merged: 2026-09-05T18:45:48Z
provenance: []
---

# Report repository health with writrun doctor

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md) · [spec-0004](../specs/spec-0004-doctor-command.md)

Report whether the repository still satisfies what the methodology
assumes, grouped by stage and judged only up to the declared one —
files at stage 1, the forge at stage 2, Issues at stage 3. Report
only; never repair.

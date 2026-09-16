---
id: task-0033
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0036, spec-0037, spec-0041]
doc_ref: product/adoption/doctor.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-13T22:29:17Z
queued: 2026-09-16T04:17:58Z
completed: null
merged: null
provenance: []
---

# Answer what each stage requires, and what the next one would

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md)

`doctor` reports only what is broken, so a requirement that holds
prints nothing. A reader cannot tell a repository that satisfies nine
checks from one this binary never examined, and cannot see what a stage
would ask of them before declaring it. `config` writes the stage with
no word about what the new stage requires.

[`adoption/doctor.excalidraw`](../../docs/product/screens/adoption/doctor.excalidraw)
and [`adoption/config.excalidraw`](../../docs/product/screens/adoption/config.excalidraw)
draw the answer: every requirement is a row marked met or not, the
stage above the declared one is previewed, each row is navigable and
explained in the words
[`adoption/doctor.md`](../../docs/product/adoption/doctor.md) carries,
and raising the stage shows those same answers before it writes.

It matters because the question a person actually asks is "can I move
up yet", and today the only answer is a silence that means yes.

---
id: task-0018
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0017]
doc_ref: product/pull-requests/shape.md
origin: report
priority: high
depends_on: []
milestone: null
created: 2026-09-05T19:11:04Z
queued: 2026-09-06T00:22:35Z
completed: 2026-09-06T00:37:20Z
merged: null
provenance: []
---

# Settle when finish writes, so a decline leaves nothing and preflight sees the edits

**References:** [product/pull-requests/shape.md](../../docs/product/pull-requests/shape.md) · [spec-0017](../specs/spec-0017-finish-write-order.md)

`shape.md` says a refused pull-request command leaves nothing behind.
`writrun finish` refuses and leaves the spec `implemented` and the
task's completion date written, because spec-0010 puts the two writes
before the confirmation — the check that follows them reads the date
they leave. The same order makes the later checks judge a range the
edits are not in.

Settle which of the two statements gives. `writrun author` and
`writrun amend` are not written yet and will inherit whichever shape
this leaves, so deciding now costs one command instead of three.

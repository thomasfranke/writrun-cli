---
id: task-0009
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0008]
doc_ref: product/pull-requests/take.md
origin: rule
priority: medium
depends_on: [task-0001]
milestone: null
created: 2026-09-03T22:30:22Z
queued: 2026-09-04T05:05:03Z
completed: 2026-09-05T17:27:46Z
merged: null
provenance: []
---

# Take a task with writrun take

**References:** [product/pull-requests/take.md](../../docs/product/pull-requests/take.md) · [spec-0008](../specs/spec-0008-take-command.md)

Start work on a task: checks first in their fixed order, branch and
title and body composed from the project's conventions, everything
shown, the draft pull request opened on confirmation. A refused take
leaves nothing behind.

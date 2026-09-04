---
id: task-0001
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0001]
doc_ref: product/rules.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-03T22:30:16Z
queued: 2026-09-04T05:05:03Z
completed: null
merged: null
provenance: []
---

# Build the command frame every command runs in

**References:** [product/rules.md](../../docs/product/rules.md) · [spec-0001](../specs/spec-0001-command-frame.md)

Every command shares one frame: run only where it may run, refuse
loudly elsewhere, confirm before changing anything, stop at the first
failing check, exit zero only on success, every question navigable by
arrow keys where stdin is a terminal and answerable by a flag
everywhere. Build that frame once — dispatch, adopted-repository
detection, the interaction helpers, `--version` and `--help` answering
anywhere — so every later task implements only its own command. Blocks
the entire queue.

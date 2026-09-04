---
id: task-0002
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0002]
doc_ref: product/adoption/init.md
origin: rule
priority: high
depends_on: [task-0001]
milestone: null
created: 2026-09-03T22:30:17Z
queued: 2026-09-04T05:05:03Z
completed: null
merged: null
provenance: []
---

# Install the WritRun kit with writrun init

**References:** [product/adoption/init.md](../../docs/product/adoption/init.md) · [spec-0002](../specs/spec-0002-init-command.md)

Adoption today is hand-copying `template/` and following `WRITRUN.md`
by eye. Make it one command: the kit installed at a pinned WritRun tag,
the repository's existing conventions extracted, an existing `AGENTS.md`
grafted, the commit-message hook installed, the stage chosen and its
requirements checked on the spot, the queue left empty.

---
id: task-0017
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0016]
doc_ref: technical/engineering/boundaries.md
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-05T13:58:35Z
queued: 2026-09-05T14:07:26Z
completed: null
merged: null
provenance: []
---

# Give the fetch a fake, as every other boundary has

**References:** [technical/engineering/boundaries.md](../../docs/technical/engineering/boundaries.md) · [spec-0016](../specs/spec-0016-the-fetch-is.md)

`boundaries.md` puts everything leaving the process behind a small
interface with a fake beside it. Script execution, `gh`, the terminal
and now the filesystem are. The fetch is the fifth and has no fake: it
clones a tag from a remote, which is as much a departure from this
process as any of them.

The cost is already recorded. `init` and `update` cannot be driven end
to end in a unit test, because the one thing between them and a fixture
is a real `git clone`; the two tests written for their partial-state
messages were removed over it, and those messages are among the
statements no fixture reaches.

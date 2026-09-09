---
id: task-0031
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0034]
doc_ref: product/screens/README.md
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-09-09T06:24:00Z
queued: 2026-09-09T06:36:46Z
completed: 2026-09-09T07:05:00Z
merged: null
provenance: []
---

# Give `config` the screen its drawing already shows

`config.excalidraw` draws a screen: a cursor on a key, a detail line
beneath, and the footer `↑↓ move · enter change · q back`. The command
prints every key and returns. So the drawing's navigation does not
exist — the maintainer found it by trying to use it, and reported it as
not being able to navigate.

`screens/README.md` now states what holds for that screen. This task is
its derived work.

The edit path is already built:
[spec-0031](../specs/spec-0031-config-screen.md) writes a value, runs
the kit's checker, and restores the previous bytes on a refusal. What
is missing is the screen in front of it — moving, and `enter` on the
selected key calling what already exists.

Nothing here waits on the kit. Offering the *allowed values* before the
write does, and that is
[report-0030](../reports/report-0030-settings-vocabulary-unreadable.md),
open and untriaged. This task asks for the value as the command already
asks for it: free text, judged after.

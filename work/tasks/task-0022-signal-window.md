---
id: task-0022
status: backlog
blocked_reason: null
taken_by: null
spec_ref: [spec-0021]
doc_ref: null
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-06T02:52:54Z
queued: null
completed: null
merged: null
provenance: []
---

# Survive a signal between finish's completion writes and its confirmation

**References:** [spec-0021](../specs/spec-0021-signal-window.md)

`writrun finish` writes the spec's `implemented` and the task's
`completed` at step 2 and undoes them on every later end that is not a
success. The undo is ordinary Go control flow, so a signal that kills
the process runs none of it: nothing in the binary imports `os/signal`.

report-0018 reproduced it — a SIGTERM two seconds into a slowed
`preflight.sh`, exit 143, and both files left changed with the spec
`implemented` and the task dated. That is report-0015's original finding
standing again on the one path the undo does not reach.

The window is not a hairline. It spans `record_provenance.sh`,
`preflight.sh` and however long the human takes at the question, and
`preflight.sh`'s own comment says the queue sweep "takes long enough
that the silence reads as a hang" — which is a description of the moment
someone presses Ctrl-C. spec-0017 named a dirty window as the price of
the order it kept and did not size it; this is the work of sizing it
down to nothing.

Ctrl-C at the confirmation prompt is already safe and is not this: huh
reads it as a key, the error travels up, and the undo runs.

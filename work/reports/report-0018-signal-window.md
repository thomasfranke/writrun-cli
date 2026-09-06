---
id: report-0018
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T01:30:20Z
triaged: null
---

# A signal between the completion writes and the confirmation leaves finish's edits in the tree

`writrun finish` writes the spec's `implemented` and the task's
`completed` at step 2 and undoes them on every end after that which is
not a success, but the undo is ordinary Go control flow: nothing in the
binary imports `os/signal`, so a signal that kills the process between
step 2 and step 5 runs none of it. Reproduced on the fixture with a
slowed `preflight.sh` and a SIGTERM two seconds into it: exit 143, and
`git status --porcelain` afterwards shows ` M work/specs/spec-0001-a-thing.md`
and ` M work/tasks/task-0001-a-thing.md`, the spec `implemented` and the
task carrying its completion date — report-0015's finding exactly, on a
path the undo does not reach. Ctrl-C at the confirmation prompt is not
this: huh reads it as a key, the error travels up and the undo runs. The
window is the whole of steps 3 to 5 — `record_provenance.sh`,
`preflight.sh`, and however long the human takes at the question — and
`preflight.sh`'s own comment says the queue sweep "takes long enough
that the silence reads as a hang". spec-0017 named a dirty window as the
cost of the order it kept and did not size it.

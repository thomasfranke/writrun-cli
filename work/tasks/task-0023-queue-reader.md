---
id: task-0023
status: done
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0022]
doc_ref: null
origin: report
priority: high
depends_on: []
milestone: null
created: 2026-09-06T04:28:17Z
queued: 2026-09-06T05:07:59Z
completed: 2026-09-06T08:00:07Z
merged: 2026-09-06T19:02:45Z
provenance: []
---

# Read the queue in one place, the way the kit reads it

**References:** [spec-0022](../specs/spec-0022-queue-reader.md)

Four packages read the queue's front matter with four copies of the
same code — `amendcmd`, `finishcmd`, `authorcmd` and `statuscmd` — and
they have already drifted apart on thirteen points. Whether a CRLF file
has front matter, whether an unclosed block has fields, whether the
first or the last of two `status:` lines wins, whether `spec_ref:
[null]` reads as empty or as one entry named `null`: each copy answers
differently, and nobody can say what "the queue's front matter" means in
this binary without reading four files.

The kit already settles every disputed point. `queue_lib.sh`'s
`ql_fm_field` matches the fence exactly, `check_front_matter.sh`'s
`fm_block` requires the closing one, and `ql_task_num` strips only the
prefix of the kind being resolved. `about.md` says that when the binary
and the scripts disagree, the scripts are right — so the answers are not
in dispute, only unapplied.

The copy that corrupted has already been fixed on its own: `finish`
resolving a spec id to a task reached the write stage silently, and that
guard landed separately because it could not wait for this. What is left
is the drift, which grows — `authorcmd` and `statuscmd` were each
written after the rule was settled and neither followed it.

---
id: report-0025
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T03:06:41Z
triaged: null
---

# take_task.sh names a resume it refuses, and never writes the commit the pull request needs

`writrun take task-0021` pushed `task/0021-queue-screen` and then failed:
`gh pr create` answered `GraphQL: No commits between main and
task/0021-queue-screen`. The branch is cut from `main` and
`take_task.sh` writes nothing onto it — the queue's status line is the
machinery's, so the script has nothing of its own to commit — and the
forge refuses a pull request across an empty range. The anchoring commit
that earlier takes carry (`c8d10c6`, "chore(queue): take task-0019,
anchoring the draft", empty) is a convention kept by whoever takes the
task, not something the script writes, so a take that is only the script
cannot succeed on a repository whose forge refuses empty ranges.

The recovery it printed does not run. `take_task.sh:438` names
`--resume --confirm` for the state it just produced — "pushed but has no
pull request, which is the one state this act must not leave behind" —
and `take_task.sh:298` refuses that same state: "already on the forge —
what --resume finishes is a branch that never reached it". The carve-out
at 293-297 is written for a local branch with no upstream, which is not
the state 438 detects. One file names a recovery and rejects it.

The draft for task-0021 was opened by hand instead, with an empty
anchoring commit; the machinery then wrote `in-progress` and `taken_by`
from the draft event as usual. `writrun author` and `writrun amend`
carry their own version of the same seam — the resume after a landed
push — recorded in this round's review as findings on both.

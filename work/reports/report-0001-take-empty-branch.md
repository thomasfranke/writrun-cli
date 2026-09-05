---
id: report-0001
status: declined
task_ref: []
doc_ref: null
created: 2026-09-04T05:22:00Z
triaged: 2026-09-05T11:55:00Z
---

# take_task.sh cannot open the draft on a commit-less branch

Taking task-0001, `take_task.sh` cut the branch from `origin/main`,
pushed it, and failed to open the draft: the forge refuses a pull
request with no commits between the branches (`GraphQL: No commits
between main and task/0001-command-frame`). The script never commits,
so every fresh take reaches the forge in that state. `--resume` then
refuses too: what it finishes is a branch that never reached the forge,
and this one had. Worked around by hand on every take since — an empty
commit, push, `gh pr create --draft`.

**Declined here: the defect is the kit's, not this CLI's.** `take_task.sh`
is `.writrun/`, which belongs to WritRun upstream, and this repository's
queue is for the `writrun` binary. It is recorded there as `REPORT-0019`
and tracked there as `TASK-0052`, "Give the take a recovery that works
after the push". Declining destroys nothing: the record stays here,
where a person can disagree with it.

**This file is a restoration.** It was written on pull request #17 and
removed from that branch before the merge, so no merged commit carries
it — while its mirror, Issue #18, kept the id. That is what let
`ql_next_id` hand `report-0001` out a second time, and what reopened a
triaged Issue under this title (`report-0031` upstream).

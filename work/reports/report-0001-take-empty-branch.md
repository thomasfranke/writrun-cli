---
id: report-0001
status: open
task_ref: []
doc_ref: null
created: 2026-09-04T05:23:19Z
triaged: null
---

# take_task.sh cannot open the draft on a commit-less branch

Taking task-0001, `take_task.sh` cut the branch from `origin/main`,
pushed it, and `gh pr create` was refused with `GraphQL: No commits
between main and task/0001-command-frame` — the forge does not open a
pull request on a branch identical to its base, and the script pushes
before any commit exists. The exit-3 message names `--resume` as the
finisher, but `--resume` refused the same state with
"task/0001-command-frame is already on the forge — what --resume
finishes is a branch that never reached it": it covers the failed-push
state only, not the pushed-but-no-PR state the script itself calls "the
one state this act must not leave behind". The act was finished by hand
with an empty commit and a manual `gh pr create --draft` (PR #17).

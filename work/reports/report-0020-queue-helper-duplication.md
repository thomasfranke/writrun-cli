---
id: report-0020
status: tracked
task_ref: [task-0023]
doc_ref: null
created: 2026-09-06T01:29:48Z
triaged: 2026-09-06T04:28:17Z
---

# The queue front-matter helpers are duplicated with no task behind them

**References:** [task-0023](../tasks/task-0023-queue-reader.md)

`numOf`, `queueFile`, `frontMatter`, `field`, `setField` and `specRefs`
stand in `internal/command/amendcmd/queue.go` and in
`internal/command/finishcmd/queue.go`, byte for byte apart from their
doc comments, and `internal/command/statuscmd/queue.go` carries a third
reading of the same front matter; `split` is now in three copies
(`takecmd.go`, `finishcmd.go`, `amendcmd.go`), each saying in its own
comment that it is kept rather than shared. spec-0011's Outcome and pull
request #62 both said task-0019 would consolidate this, but spec-0018 is
about `.writrun/VERSION` readers and the stage-0 requirement list:
`gh pr view 64 --json files` lists `internal/kittag/`,
`internal/requirements/`, `doctorcmd`, `initcmd`, `statuscmd`,
`updatecmd` and `docs/technical/layout/tree.md`, and touches neither
`finishcmd` nor `amendcmd`. The sentence has been removed from
spec-0011's Outcome, so the duplication now has nothing pointing at it.

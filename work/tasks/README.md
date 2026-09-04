# Tasks — the queue

One file per task, named `task-NNNN-<subject>.md` — a four-digit id and
an extremely short subject slug (`task-0005-multi-file-search.md`), so a
directory listing reads as a queue summary. Never renamed, never
moved. Status, priority, and dependencies live in front-matter — schema
and selection algorithm are defined in WritRun's technical README
(`docs/technical/README.md#task-schema` in the WritRun repository).

A task is the **request**: what to do, when, what blocks it. No technical
detail — scope, steps, and acceptance criteria belong to its spec(s) in
[`../specs/`](../specs/README.md). The task always exists before its specs.

Trivial work is a commit, never a task.

## For agents

Do not choose work by reading this README or by directory listing order.
Run the selection algorithm (`writrun-select-next-task` skill): resume
`in-progress` first, then filter, sort, and take the first. Read every
referenced spec and doc anchor before writing code.

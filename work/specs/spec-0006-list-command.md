---
id: spec-0006
task_ref: task-0006
status: implemented
created: 2026-09-03T22:30:38Z
---

# spec-0006 — Render the queue with eligibility unchanged

**References:** [task-0006](../tasks/task-0006-list-command.md)

- **Goal:** `writrun list` renders the queue; eligibility is the methodology's, unchanged.

## Scope

In: running the selection skill and presenting its sections; filters
that narrow the listing. Out: any change to how eligibility or order is
decided; any write.

## Steps

1. Run `.writrun/skills/writrun-select-next-task/list_tasks.sh` from the repository root; stream its sections.
2. Filters select sections (`--available`, `--held`, `--reports`); they never change a task's group or order.
3. Map exit codes: something available → 0; nothing → 0 with the script's own message; no queue directory → abort naming the cause.

## Acceptance criteria (EARS)

- When the queue directory is missing, the system shall abort naming the cause.
- When a filter is given, every task shown shall be in the same group and order as the unfiltered run.
- When the forge is unreachable, the system shall answer from the files and pass the script's warning through.
- When run, the system shall write nothing.

## Edge cases

- Empty queue: the script's `Nothing is available.` passes through.
- A held-back reason line: shown verbatim, never rephrased.

## Tests required

Integration over queue fixtures: empty, mixed statuses, forge stubbed
reachable and unreachable.

## Definition of Done

- [x] Output matches the skill's for every fixture, modulo filtering.
- [x] Suite green.

## Proposed product changes

- none — `product/queue/list.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

`writrun list` is `internal/command/listcmd/`, registered in
`cmd/writrun/main.go` and needing an adopted repository. It runs
`.writrun/skills/writrun-select-next-task/list_tasks.sh` from `ctx.Root`
through the exec port — `listcmd.Script`, wired to `kit.Run`, which is
`kit.Runner` for a caller holding the root and the streams.

- Unfiltered, the script's stdout is the command's: the port writes
  straight to the user's terminal and no byte is read or rewritten.
- A filter buffers the output and `sections.go` cuts it into blocks at
  the lister's own headings. A block is printed whole or not at all, in
  the order the lister printed it, so no task changes group or order.
  `--available` selects the in-progress, available and in-flight
  sections, `--held` the held-back section, `--reports` the open
  reports; the lister's notes and anything preceding its first heading
  print in every run.
- Exit 1 from the script becomes exit 0 — nothing available is an
  answer, and `Nothing is available.` is the message that carries it.
  Every other non-zero exit travels up with its code, so a missing
  `work/tasks/` exits 3 with the script's `No such directory:` on
  stderr.

Tests: `internal/command/listcmd/` unit tests over a faked port (23
cases), and `tests/integration/list/` over `tests/list_lib.sh` — a
repository carrying this repository's own copy of the lister, a queue
per case, and `gh` stubbed (28 checks). The mixed-status case diffs the
binary's stdout against a direct run of the lister and requires them
identical. Coverage: `internal/command/listcmd` 98.6%, total over
`internal/` 98.1%.

No permanent doc changed: `product/queue/list.md` already stated the
behaviour, and nothing was found to correct in it.

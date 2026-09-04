---
id: spec-0006
task_ref: task-0006
status: draft
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

- [ ] Output matches the skill's for every fixture, modulo filtering.
- [ ] Suite green.

## Proposed product changes

- none — `product/queue/list.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

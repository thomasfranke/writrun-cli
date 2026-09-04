---
id: spec-0013
task_ref: task-0014
status: approved
created: 2026-09-03T23:27:24Z
---

# spec-0013 — Read the branch, the checks and the queue; write nothing

**References:** [task-0014](../tasks/task-0014-status-command.md)

- **Goal:** `writrun status` reads where the work stands; it writes nothing.

## Scope

In: resolving the branch to a task, the read-only run of the completion
checks, the open-report count, the kit-tag comparison. Out: any write;
any repair; any judgement the checks do not already make.

## Steps

1. Resolve the current branch to a task by the `task/NNNN-*` name; no task — say so and skip step 3.
2. Name the task's spec and the spec's status from the queue files.
3. Run the completion checks (`preflight.sh`) read-only; name the first that fails — what `finish` would stop at — or that all pass.
4. Count the reports with status `open` in `work/reports/`.
5. Compare `.writrun/VERSION` with the tag this client pins; name a mismatch, bridge nothing.

## Acceptance criteria (EARS)

- When run, the system shall write nothing.
- When the branch carries no task, the system shall say so and still report the open reports and the kit tag.
- When a completion check fails, the system shall name that check and its failure.
- When `.writrun/VERSION` differs from the pinned tag, the system shall name both values.

## Edge cases

- Detached HEAD, or `main`: the no-task path, said plainly.
- A branch named like a task the queue does not hold: named as unknown, never invented.

## Tests required

Integration over fixture repositories: with and without a task branch,
passing and failing checks, matching and mismatched kit tags.

## Definition of Done

- [ ] Every acceptance criterion has a test.
- [ ] Suite green.

## Proposed product changes

- none — `product/queue/status.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

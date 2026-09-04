---
id: spec-0010
task_ref: task-0011
status: approved
created: 2026-09-03T22:30:42Z
---

# spec-0010 — Verify promises, record outcome, mark ready

**References:** [task-0011](../tasks/task-0011-finish-command.md)

- **Goal:** `writrun finish` closes the loop in the load-bearing order.

## Scope

In: the completion sequence — deltas verified, outcome recorded, spec
`implemented`, `completed` written, state checked, PR marked ready. Out:
the task's `status` line (the machinery's, always); the merge.

## Steps

1. Run `check_deltas.sh` for the branch's specs; stop on non-zero — nothing else runs.
2. Require the spec's `## Outcome` filled; write `status: implemented` on the spec and the task's `completed` UTC timestamp. Touch the task's `status` line never.
3. Run `record_provenance.sh` unconditionally — it reads the ledger setting itself.
4. Run `preflight.sh`; exit 0 required.
5. Show what will happen; on confirmation mark the pull request ready for review.

## Acceptance criteria (EARS)

- When the delta check fails, the system shall stop before writing any status.
- When the spec's Outcome is empty, the system shall refuse to mark it implemented.
- When finish completes, the task's `status` line shall be unchanged.
- When preflight is non-zero, the pull request shall not be marked ready.

## Edge cases

- A task with empty `spec_ref`: no deltas to check; `completed` is still written; preflight still gates.
- Several specs on one branch: one `check_deltas.sh` call carrying all of them.

## Tests required

Integration over a fixture adoption: green path, each failing gate,
multi-spec branch.

## Definition of Done

- [ ] The order of operations matches `finishing.md`'s and is asserted by tests.
- [ ] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

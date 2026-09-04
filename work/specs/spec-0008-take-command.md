---
id: spec-0008
task_ref: task-0009
status: draft
created: 2026-09-03T22:30:40Z
---

# spec-0008 — Checks in order, composition shown, draft opened on confirmation

**References:** [task-0009](../tasks/task-0009-take-command.md)

- **Goal:** `writrun take` wraps the methodology's own take, adding the human's confirmation.

## Scope

In: wrapping `take_task.sh` — eligibility, composition, the draft PR —
and the confirmation flow around its conduct gate. Out: writing any
queue file (the machinery answers the draft); any check reimplemented.

## Steps

1. Collect the task id — given as an argument, or arrow-selected from the available tasks where stdin is a terminal — and the title (`--title`, or typed; the one free-text question). Validate nothing locally — the script is the authority.
2. Run `take_task.sh <id> --title <t> [--slug s]`; map its exits:
   - 0 — done: branch pushed, draft open; report it.
   - 1 — refusal: pass the script's reason through, exit non-zero.
   - 2 — composed and waiting: show the composition, ask; on yes re-run with `--confirm`; `--yes` answers for the user.
   - 3 — git/forge failure: pass the reason and the `--resume` command through.

## Acceptance criteria (EARS)

- When no task id is given and stdin is a terminal, the system shall offer the available tasks for arrow selection.
- When any eligibility filter refuses the task, the system shall report the script's reason and create nothing.
- When the user declines the shown composition, nothing shall reach the forge.
- When the script exits 3 after the branch was cut, the system shall show the exact `--resume` invocation.
- When the take completes, the task file shall be unchanged in the working tree.

## Edge cases

- `--yes` with conduct flags false: the composition is still printed before the confirmed re-run.
- Title refused by the declared style: the script's refusal passes through verbatim.

## Tests required

Integration with the kit's `WRITRUN_PR_LIST` / `WRITRUN_PR_FILES` seams
and a stubbed `gh`.

## Definition of Done

- [ ] Every script exit code has a mapped, tested behaviour.
- [ ] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

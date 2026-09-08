---
id: spec-0030
task_ref: task-0029
status: approved
created: 2026-09-08T15:04:02Z
---

# spec-0030 — doctor examines a stage above the declared one

**References:** [task-0029](../tasks/task-0029-ui-screens-refactor.md)

- **Goal:** A reader can ask what a stage they have not declared would
  require of them, and get the answer `doctor` already knows how to
  compute.

## Scope

In: examining a named stage rather than only the declared one, and
saying plainly that the report is hypothetical.

Out: changing a check, changing what any level means, and writing the
declared stage — the settings are
[spec-0031](spec-0031-config-screen.md)'s.

## Steps

1. Add the flag that names the stage to examine. Without it, nothing
   changes: stages 0 to the declared one, as
   [`doctor.md`](../../docs/product/adoption/doctor.md) states.
2. Run the checks for stages 0 to the named one, through the code paths
   the declared run already uses. No check is duplicated for the
   hypothetical case.
3. Head the report with the stage examined and the stage declared, so a
   run above the declaration cannot be mistaken for the repository's
   real verdict.
4. Keep the exit status reading the declared stage's findings alone. A
   stage nobody declared cannot fail a build.

## Acceptance criteria (EARS)

- When `doctor` runs without the flag, the system shall examine stages 0
  to the declared one and print what it prints today, unchanged.
- When `doctor` runs with the flag naming a stage above the declared
  one, the system shall examine stages 0 to the named one.
- When a stage above the declaration is examined, the system shall name
  both the examined stage and the declared one in the first line.
- When a stage above the declaration is examined, the system shall base
  the exit status on the declared stage's findings alone.
- When the flag names a stage at or below the declared one, the system
  shall examine to that stage and report as it would for a declaration
  of it.
- When the flag names a value outside 1–3, the system shall refuse
  naming the accepted values, and shall examine nothing.
- When a check the named stage needs cannot reach the forge, the system
  shall report it `unread`, which never fails a run.

## Edge cases

- **A repository at stage 1 with no `gh` and no remote.** Every stage-2
  and stage-3 check answers `unread`, which is the honest report: the
  requirements are unknown, not unmet.
- **The named stage equals the declared one.** The output is the
  ordinary run; the header names one stage twice rather than inventing a
  second form.
- **The declared stage is unreadable.** `doctor` already assumes stage 1
  and names why; the flag still governs what is examined, and the header
  names the assumption.

## Tests required

- Unit, `internal/command/doctorcmd`: the examined range follows the
  flag; the exit status follows the declaration; the header names both.
- Unit: an out-of-range value refuses and examines nothing.
- Integration, `tests/integration/doctor/`: a stage-1 fixture examined at
  stage 2 reports the forge checks rather than `not examined`, and exits
  on the stage-1 findings alone.

## Definition of Done

- [ ] A stage above the declaration can be examined.
- [ ] No check is written twice for the hypothetical case.
- [ ] The exit status still answers for the declared stage only.
- [ ] `doctor.md` states the flag and what it does not change.

## Proposed product changes

- `product/adoption/doctor.md` — the flag, the header naming both
  stages, and the rule that the exit status answers for the declared
  stage alone.
- `product/screens/` — the doctor drawing gains a frame showing a stage
  examined above the declaration, replacing the note that says doctor
  declines the question.

## Proposed technical changes

- none — no machinery change.

## Outcome

_(fill after execution)_

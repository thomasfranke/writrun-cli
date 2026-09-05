---
id: spec-0008
task_ref: task-0009
status: implemented
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

Shipped as specified. `writrun take` lives in
`internal/command/takecmd/`: the task id is the argument or the
lister's Available group arrow-selected, the title is `--title` or the
one free-text question, and
`.writrun/scripts/stage-2-pull-requests/take_task.sh` decides
everything else. The four exit codes map as the steps say — 0 returns
having added nothing to the script's report, 1 and 3 return the error
carrying the script's own code so the frame passes it through
unedited, 2 asks and reruns with `--confirm`.

Three decisions the spec did not name:

- **The frame gained the free-text question.** `Terminal.Input` and
  `Ctx.AskInput` are the typed answer `product/rules.md` allows;
  `--yes` does not answer it, because a value nobody wrote is not an
  answer a flag can stand in for.
- **`internal/kit` became a `Runner` function**, shaped like
  `gitx.Runner`, with the streams as arguments: take shows the take
  script's reporting and reads the lister's back for the selection, and
  one port serves both.
- **A confirmed rerun that exits 2 again is named**, not passed
  through: it would leave the composition printed twice and the act
  undone with nobody saying so.

Verified by 21 unit tests over the faked runner and 20 assertions in
`tests/integration/take/` — seven cases over the real `take_task.sh`, a
bare origin and a stubbed `gh`, with `WRITRUN_PR_LIST` and
`WRITRUN_PR_FILES` supplying the open pull requests, one case per exit
code plus the in-flight and amendment refusals. `make tests` exits 0
(55 case files passed, 0 failed); `make cover` exits 0 at 98.1% over
`internal/`, `takecmd` at 97.6%;
`writrun-check-spec-deltas spec-0008` exits 0.

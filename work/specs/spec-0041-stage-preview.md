---
id: spec-0041
task_ref: task-0033
status: draft
created: 2026-09-13T22:08:11Z
---

# spec-0041 — Raising the stage previews what the stage requires

**References:** [task-0033](../tasks/task-0033-stage-requirements.md)

- **Goal:** Raising the stage shows what the new stage would require of this repository, then asks.

## Scope

In: the preview `config` runs when the stage rises, and the question it
ends in.

Out: any refusal. The stage stays a declaration, and this spec adds no
gate to it.

## Steps

1. When the stage is raised, run the target stage's checks through
   `doctorcmd`'s own port. `config` holds no second copy of a check.
2. Show them with the marks and counts
   [spec-0036](spec-0036-doctor-requirements.md) gives the report.
3. Ask. `enter` writes the stage, `esc` cancels, `d` opens the doctor
   screen at that stage.
4. Never refuse. A check that could not be read is `unread`, and an
   adopter with no network can still declare a stage.
5. Run no preview when the stage is lowered or unchanged: nothing new is
   required of the repository.
6. Without a terminal, write as today and print the preview rather than
   asking it.

## Acceptance criteria (EARS)

- When the stage is raised on a terminal, the system shall show the
  target stage's requirements before writing.
- When a previewed check cannot be read, the system shall mark it
  `unread` and still offer the write.
- When the reader confirms, the system shall write the stage and hand
  the file to the kit's checker, as it does today.
- When the stage is lowered or unchanged, the system shall run no
  preview.
- When there is no terminal, the system shall write the value and print
  the preview.

## Edge cases

- **Raising to a stage whose checks all answer `unread`.** The count
  says so, and the write is still offered.
- **A refused write.** The checker's own sentence stands and the file is
  restored, unchanged by this spec.
- **Raising from 1 to 3.** The preview covers the target stage, not each
  rung between.

## Tests required

Unit, `internal/command/configcmd`: the preview runs on a raise and not
on a lower; an unreadable check does not block the write.

Integration, `tests/integration/cli/`: a stage raise with a stubbed `gh`
shows the requirements and writes on confirmation.

The drawing is the assertion. Each frame named above is read out of its
`.excalidraw` and compared with what the binary renders, and a
disagreement fails the suite. The drawing is not the thing under test:
it is what the test is written against.

## Definition of Done

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [ ] The preview is doctor's answers, not a second implementation.
- [ ] No path refuses a declaration.

## Proposed product changes

- [`product/config.md`](../../docs/product/config.md) — the preview and
  what it never does.

## Proposed technical changes

- `internal/command/configcmd` — the call into `doctorcmd`'s checks, and
  the port that makes it possible without a cycle.

## Outcome

_(fill after execution)_

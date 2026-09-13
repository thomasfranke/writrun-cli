---
id: spec-0038
task_ref: task-0034
status: draft
created: 2026-09-13T22:08:05Z
---

# spec-0038 — Open a screen where the kit is absent

**References:** [task-0034](../tasks/task-0034-screen-furniture.md)

- **Goal:** Where `.writrun/` is absent, `writrun` opens a screen that says what this is, what the environment answers, and what to do next.

## Scope

In: the screen, the environment marked with `doctor`'s own marks, and
`init` held out of reach while a requirement is unmet.

Out: what `init` itself does; the entry screen inside an adoption.

## Steps

1. Where `.writrun/` is absent and both ends of the terminal are
   terminals, open the screen instead of printing what `--help` prints.
2. Head it with the wordmark and WritRun's own line — `What is written,
   runs` — then the identity line and a context line saying no kit is
   here. The wordmark appears here and on no other screen.
3. Answer the four environment requirements with the same probe
   `doctor`'s stage 0 uses (`internal/requirements`), marked the same
   way.
4. Offer `init` as the one adoption row, and hold it out of reach while
   any requirement is unmet: the row says why, and the cursor rests on
   the requirement that put it there.
5. Offer `r` to read the `PATH` again, and drop `enter` from the footer
   while nothing can run.
6. Without a terminal, print what `--help` prints, exactly as today.

## Acceptance criteria (EARS)

- When `.writrun/` is absent and a terminal is present, the system shall
  open the first-run screen.
- When every environment requirement is met, `enter` on `init` shall run
  `init` with its own questions and its own confirmation.
- When a requirement is unmet, the system shall not run `init`, shall
  say why on the row, and shall name what is missing in the footer.
- When stdin or stdout is not a terminal, the system shall print what
  `--help` prints and open nothing.
- When the directory is not a git repository, the system shall still
  open, and `init`'s own refusal shall be the answer.
- When `.writrun/` is present, the system shall open the entry screen.

## Edge cases

- **A repository that is not a git repository.** Drawn: the screen
  opens, and the refusal is the command's.
- **A terminal shorter than the screen.** The list scrolls; the footer
  stays.
- **`awk` on the `PATH` but not POSIX.** Out of scope: the probe answers
  presence, as `doctor`'s does.

## Tests required

Unit, `internal/screen`: the gate on the environment; the row's
sentence; the footer's keys with nothing runnable.

Integration, `tests/integration/screen/`: a directory with no kit and a
fake terminal opens the screen; the same without a terminal prints the
help.

The drawing is the assertion. Each frame named above is read out of its
`.excalidraw` and compared with what the binary renders, and a
disagreement fails the suite. The drawing is not the thing under test:
it is what the test is written against.

## Definition of Done

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [ ] The screen opens where the kit is absent, and only there.
- [ ] `init` is unreachable while any requirement is unmet.
- [ ] The no-terminal path is unchanged.

## Proposed product changes

- [`product/screens/README.md`](../../docs/product/screens/README.md) —
  the rule that outside an adoption the binary prints the help, replaced
  by the screen.
- [`product/rules.md`](../../docs/product/rules.md) — where a command
  runs.

## Proposed technical changes

- none — no machinery change.

## Outcome

_(fill after execution)_

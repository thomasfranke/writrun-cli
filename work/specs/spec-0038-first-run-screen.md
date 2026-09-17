---
id: spec-0038
task_ref: task-0034
status: implemented
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

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`first-run.excalidraw`](../../docs/product/screens/first-run.excalidraw)
  — `writrun — a repository with no kit in it` and `writrun — the
  environment short, init out of reach`.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [x] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [x] The screen opens where the kit is absent, and only there.
- [x] `init` is unreachable while any requirement is unmet.
- [x] The no-terminal path is unchanged.

## Proposed product changes

- `product/screens/README.md` — the rule that outside an adoption the
  binary prints the help, replaced by the screen.
- `product/rules.md` — where a command runs.

## Proposed technical changes

- none — no machinery change.

## Outcome

Built. `internal/screen/firstrun.go` is the screen: the wordmark and
WritRun's own line, the identity and context lines every other screen
opens with, the four environment requirements answered by
`internal/requirements` and marked with `doctor`'s own glyphs, `init`
as the one adoption row, and `--version` and `--help` under it.
`command.Frame` gained a `FirstRun` port beside `Screen`, wired in
`cmd/writrun/main.go`, and `openScreen` now asks the terminal first and
the kit second: a terminal at both ends opens a screen, and which
screen is whether `.writrun/` is there.

`enter` is refused while any requirement is unmet — the screen's own
gate, and the row is emptied of its command besides — and the footer
offers `r re-check` in `enter`'s place. `r` reads the `PATH` again
through the same probe, so a requirement installed while the screen is
open is met the moment it is asked for.

Both frames are asserted line for line
(`internal/screen/screens_frames_test.go`), and the screen is opened on
a pty in `tests/e2e/screen/the_first_run_screen_test.sh`. The
no-terminal path is unchanged and proved so in
`tests/integration/screen/the_screen_needs_a_terminal_test.sh`.

**What the plan did not foresee.**

The command a key chooses runs **after** the screen closes, not inside
it. The session pauses for a command because it has more to offer
afterwards; this screen's whole offer is one command, so closing it is
enough — and the spawn port is still used where it is wired, for the
reason decision 0015 gives.

The four requirements needed a sentence each, and `doctor`'s are the
document's table, which is not readable where no repository has been
adopted. `requirements.Reason` carries one clause per binary beside the
list that names them — the same fact, not a second one — and the screen
composes the rest.

`init`'s row shows the command table's own summary rather than a line
written for this screen. The drawing stated a third wording of one
field, which is what report-0042 recorded; the row now reads from the
table like every other row on every other screen.

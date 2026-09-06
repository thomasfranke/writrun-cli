---
id: spec-0020
task_ref: task-0021
status: approved
created: 2026-09-06T02:53:08Z
---

# spec-0020 — The no-command queue screen

**References:** [task-0021](../tasks/task-0021-queue-screen.md)

- **Goal:** `writrun` with no command opens the queue as a key-navigated screen, and every key leaves the screen and runs the command it names, unchanged.

## Scope

In: the no-command path in `cmd/writrun/main.go`; a screen package
rendering the sections `list` already renders; the dispatch from a key
to an existing command. Out: every command's own behaviour — none is
modified, and the screen adds no action any of them lacks; the output
shape of `list` itself; `--help`, which the screen falls back to.

## What is wrong

`product/screen.md` specifies the no-command entry and nothing
implements it. `writrun` with no arguments prints the help text on every
path, including the one the rule describes: a terminal, inside an
adopted repository. The rule's two exceptions are today its only
behaviour, so a reader of the docs and a user of the binary disagree
about what the product is.

The engine is not the obstacle. Decision
`runtime/0009-interaction-via-charm-huh.md` chose the Charm stack behind
the `term` port for the confirmations and the one free-text prompt, and
says so in as many words: "if screens ever appear, the engine is already
present". It also records that numbered menus were rejected because the
product rule keeps arrow keys — a rejection written for this screen,
before it existed.

## Steps

1. Route the no-command path: with a terminal on stdin and stdout and inside an adopted repository, open the screen; otherwise print what `--help` prints, which is what happens today.
2. Render the sections `list` renders, in `list`'s order, reading through the same queue reader so the two cannot drift into two answers about one queue.
3. Move the selection with the arrow keys, over the selectable rows only — a section heading and a held-back entry are shown, not selected.
4. Dispatch: `Enter` runs `take` on the selection, `w` runs `work` on it, `s` runs `status`, `q` leaves. The screen closes before the command runs, so the command owns the terminal it asks its questions on.
5. Leave every dispatched command untouched: its checks, its questions and its confirmation run as they do from the shell, and its exit code is the process's.

## Acceptance criteria (EARS)

- When `writrun` is run with no command, with a terminal and inside an adopted repository, the system shall open the screen.
- When `writrun` is run with no command and stdin or stdout is not a terminal, the system shall print what `--help` prints and exit 0.
- When `writrun` is run with no command outside an adopted repository, the system shall print what `--help` prints and exit 0.
- While the screen is open, the system shall write nothing to the repository.
- When a dispatch key is pressed, the system shall leave the screen before the named command begins, and shall carry that command's exit code out of the process.
- When `q` is pressed, the system shall leave the screen and exit 0 without running any command.
- When the queue holds no selectable row, the system shall open, say so, and accept `s` and `q`.

## Edge cases

- A terminal too short for the queue: the screen scrolls the selection into view rather than truncating the list silently.
- The selection is a task that is not `ready`: `Enter` dispatches `take` anyway and `take`'s own refusal is the answer — the screen judges no task.
- `w` on a repository with no configured agent: `work`'s own error, unchanged.
- The terminal is resized while the screen is open: the render follows, which is Bubble Tea's to handle and not this screen's to re-implement.
- A queue file the reader refuses: the screen reports what `list` reports for it rather than failing to open.

## Tests required

Unit over the screen's model: the key table, the selection skipping
non-selectable rows, and each key resolving to the command it names with
the selected id. Integration for the routing — no terminal prints help,
outside an adoption prints help — and one case asserting the screen
writes nothing, driven by `git status --porcelain` after opening and
leaving with `q`. The dispatched commands keep their own suites; nothing
here re-tests them.

## Definition of Done

- [ ] `writrun` inside this repository, in a terminal, opens the queue and `q` leaves it.
- [ ] `writrun | cat` prints what `--help` prints, proven by a test.
- [ ] Opening the screen and leaving it changes no file, proven by a test.
- [ ] No command's behaviour or output changed, proven by the existing suites passing untouched.

## Proposed product changes

- none — `product/screen.md` already states the rule this implements.

## Proposed technical changes

- `technical/layout/tree.md` — a row for the screen package.

## Outcome

**The screen is `internal/screen`, and the routing is the frame's.**
`Parse` turns the lister's output into rows and keeps every line
verbatim; the model holds the five keys and the selection; `Open` runs
the program and returns the action. `cmd/writrun` wires the port to
`.writrun/skills/writrun-select-next-task/list_tasks.sh` — the path
`listcmd` already names, so the two commands cannot become two answers
about one queue.

**Every task row is selectable, including one that cannot be taken.**
This spec said two things and they disagreed: step 3 called a held-back
entry unselectable, and Edge cases gave `take`'s own refusal as the
answer for a task that is not `ready`. The Edge case won. "The screen
offers no action a command does not already provide" cuts both ways —
it must not withhold one either — and a screen that hid a row would be
judging a task, which `screen.md` reserves to the command. The reason
is at `screen.Row`, where the next reader meets it.

**Adoption is read, not enforced.** `NeedAdopted` prints a refusal, and
the rule asks for the help; so the no-command path calls `FindRepo`
itself and falls back to `help`. That is the one caller that wants the
fact without the verdict, and it is said at `openScreen`.

**The screen closes before the command runs.** The action is a value
returned rather than a call made from inside the model, because a huh
form rendering underneath a live Bubble Tea program is two programs
holding one terminal. The dispatched command's exit code is the
process's; `TestTheDispatchedCommandsCodeIsTheProcesss` pins it.

**`Frame.Screen` is nil-able**, and nil prints the help — a binary
built without a screen behaves as this one did before there was one.

Not done here: no Makefile target launches the screen. The target's row
would have to change `technical/layout/makefile.md`, which this spec
did not promise, and `check_promised_deltas.sh` treats a change moving
no spec to `implemented` as having no delta to check — so it belongs to
a trivial branch of its own, not to this one.

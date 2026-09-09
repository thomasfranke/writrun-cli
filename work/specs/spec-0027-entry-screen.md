---
id: spec-0027
task_ref: task-0029
status: implemented
created: 2026-09-08T15:03:49Z
---

# spec-0027 — The entry screen lists what the binary can do

**References:** [task-0029](../tasks/task-0029-ui-screens-refactor.md)

- **Goal:** `writrun` with no command opens a screen naming every
  command that can run in an adopted repository, and the queue becomes
  the screen `list` opens from it.

## Scope

In: the entry screen's content and keys, the queue screen gaining a way
back, and the command summaries the entry screen prints.

Out: colour ([spec-0028](spec-0028-screen-colour.md)), the binary's name
([spec-0029](spec-0029-binary-name.md)), `doctor` above the declared
stage ([spec-0030](spec-0030-doctor-above-stage.md)), and the config
screen ([spec-0031](spec-0031-config-screen.md)). The entry screen's
stage block reads `.writrun/settings.json` and writes nothing.

## Steps

1. Shorten every `command.Summary` to fit a menu row — 44 characters or
   fewer, the width the drawing leaves after the name column. `--help`
   prints the same strings, so this is one edit read in two places
   ([decision 0010](../../docs/technical/decisions/docs/0010-help-is-one-line-per-command.md)).
2. Build the entry screen's rows: the header, the stage block, the four
   groups, and the detail line, laid out as
   [`entry.excalidraw`](../../docs/product/screens/entry.excalidraw)
   draws them.
3. Read the header from cheap sources only: the binary's version, the
   pinned WritRun tag, and the current branch. Read the stage block from
   `.writrun/settings.json` through the kit's own
   `read_setting.sh`.
4. Dispatch on `enter`: `list` opens the queue screen, every other entry
   leaves the screen and runs its command, unchanged.
5. Give the queue screen `esc`, which returns to the entry screen. Its
   other keys are unchanged.
6. Restate the rules in
   [`screens/README.md`](../../docs/product/screens/README.md): the
   entry screen's own section, and the queue's section beneath it.

## Acceptance criteria (EARS)

- When `writrun` runs with no command in an adopted repository on a
  terminal, the system shall open the entry screen.
- When the entry screen opens, the system shall list exactly the
  commands whose `Need` is satisfiable in an adopted repository —
  eleven today, `init` excluded because it refuses where `.writrun/`
  exists.
- When a command is added to the command table, the system shall list it
  without a second list being edited.
- When the entry screen renders a row, the system shall print that
  command's `Summary` and no string written for the screen alone.
- When a row is selected, the system shall print that command's whole
  `Summary` on the detail line.
- When `enter` is pressed on `list`, the system shall open the queue
  screen rather than run a command.
- When `enter` is pressed on any other row, the system shall close the
  screen and run that command with its own checks, questions and
  confirmation.
- When `esc` is pressed on the queue screen, the system shall return to
  the entry screen having changed nothing.
- When `q` is pressed on either screen, the system shall leave and run
  nothing.
- When stdin or stdout is not a terminal, or the repository is not
  adopted, the system shall print what `--help` prints, as today.
- When the entry screen opens, the system shall run no check and no
  script beyond reading the settings file.

## Edge cases

- **A command whose summary exceeds the row width.** The row is the
  width the drawing gives it; a longer summary is a defect in the
  summary, and step 1 is where it is fixed. The screen truncates
  nothing.
- **`.writrun/settings.json` absent or unreadable.** The kit's reader
  documents its defaults and keeps working without the file, so the
  stage block prints those defaults. It never blocks the screen.
- **A stage the settings declare outside 1–3.** `doctor` already names
  this and assumes stage 1; the block prints what the reader returned
  and adds no second opinion.
- **A terminal shorter than the screen.** The existing viewport scrolls
  the selection into view rather than truncating the list, and the
  entry screen keeps that behaviour.
- **The queue is empty.** The entry screen is unaffected — it lists
  commands, not tasks — and `list` opens a queue screen that says so.

## Tests required

- Unit, `internal/screen`: the row set is derived from the command
  table; a command added to the table appears; `init` does not.
- Unit, `internal/screen`: `enter` on `list` yields the queue screen and
  not an action; `enter` elsewhere yields that command's action; `esc`
  from the queue yields the entry screen; `q` yields the zero action.
- Unit, `internal/screen`: the detail line is the selected row's whole
  summary.
- Unit, `internal/command`: every `Summary` is within the row width, so
  the drawing and the table cannot drift apart unnoticed.
- Integration, `tests/integration/screen/`: the two no-screen cases
  still print what `--help` prints.
- Unit, `internal/palette`: a palette that was told not to paint hands
  every role's text back unchanged, and one that may paint wraps the
  text without replacing it.
- Unit, `internal/screen`: the session's states — the queue opens and
  closes, a choice on either screen is asked for and cleared, `q` ends
  it from both, the queue is re-read after a command and only while it
  is showing, and a lister that fails is named rather than fatal. And
  `waitForLine` leaves behind what was typed after the line.
- e2e, `tests/e2e/screen/`: a real pty, driven through `expect` — the
  alternate screen is asked for, the groups and footer are drawn, the
  cursor moves with the detail line following, `enter` on `list` opens
  the queue, `esc` returns to the entry screen, a command takes the
  screen and hands it back, and `q` leaves with zero. This is the only
  tier that sees the terminal rather than the model — the session is
  not provable without one — and it reads the queue without touching
  it.

## Definition of Done

- [ ] The entry screen matches
      [`entry.excalidraw`](../../docs/product/screens/entry.excalidraw)
      in rows, groups, order and keys.
- [ ] The queue screen matches
      [`queue.excalidraw`](../../docs/product/screens/queue.excalidraw),
      `esc` included.
- [ ] No string the screen prints is written twice in the source.
- [ ] `screens/README.md` states both screens' rules, and states no
      layout the drawings carry.
- [ ] Both drawings drop the `Proposed` label, in the table and on the
      canvas.

## Proposed product changes

- `product/screens/README.md` — the entry screen gains its own
  section; the existing section becomes the queue's; the table drops
  `Proposed` from Entry.
- `product/screens/` — the Entry and Queue drawings drop the
  `proposed` wording and stand as the reference; `esc` is no longer
  labelled a proposal.
- `product/README.md` — the flow table is regrouped to the screen's
  groups, so one set is not grouped two ways.
- `product/rules.md` — the no-command sentence names the entry
  screen rather than the queue.

## Proposed technical changes

- `technical/layout/tree.md` — `internal/screen` gains the entry
  screen beside the queue's rows.

## Outcome

Built as planned. The entry screen lists every command that can run
inside an adoption, grouped as the drawing groups them, with the queue
one `enter` in and `esc` back.

**The summaries were the load-bearing step.** All twelve now fit the
row, the longest having been 88 characters, and `cmd/writrun` carries a
test that fails when one outgrows it. That test earned itself inside
this same task: `config`, written hours later under
[spec-0031](spec-0031-config-screen.md), arrived two characters over and
the guard named it. It lives beside the command table rather than in
`internal/command`, because a test there would have to invent a table
and would then be guarding its own invention.

**The queue enters by callback, not by path.** `screen.Open` takes a
function that returns the lister's output, so the screen package still
knows no script path.

**A session was built and taken out again, and the rule it broke was
already written here.** With twelve commands, a screen that ends in the
shell after one is a launcher for one — the maintainer opened `config`,
read it, and was back at the prompt, and `report` had no way out of its
questions but to kill the binary. So the screen was made to return after
every command, with a `Pause` on the terminal port to keep the output
readable.

It broke six commands in six ways, all one defect. `status` hung, `doctor`
locked the binary, `config` would not navigate, `report` could not be
cancelled, `take` worked by accident, and `finish` reached its
confirmation and answered it — the `enter` that dispatched it from the
menu arriving as the answer. **Two terminal programs in one process do
not share a keyboard**, which is why the original closed the screen
*before* the command ran and said so: "a huh form rendering underneath a
live Bubble Tea program is two programs holding one terminal". Pausing
instead of leaving put a screen back on the input while a command asked
its questions.

No forge act followed — the draft was still a draft — but a confirmation
that answers itself is the failure to fear, not the one that happened.
The session is reverted, and `screens/README.md` was made to say why:
the screen is gone before the command asks anything, and does not come
back. *That rule stood for a day and is superseded — see the reversal
below.* The session is still the right shape: one Bubble Tea program for
the whole session, releasing and restoring the terminal around each
command, which is design rather than adjustment.

**None of the three UI defects in this task was found by a test**, and
that is a property of the suite rather than of these three. Every case
here drives the model — keys in, actions out — and a model answers
correctly while the terminal underneath it does not.

That gap is now closed. `tests/e2e/screen/` opens a real pty through
`expect` and reads what a person would see: the alternate screen is
asked for, the groups and footer are drawn, the cursor moves and the
detail line follows, `enter` on `list` opens the queue, `esc` returns
rather than leaving, and `q` leaves with nothing. It was held against
each defect before being trusted: with the alternate screen removed it
fails naming that, and with `esc` dispatching instead of returning it
fails naming that. It reads only — a case that took a task would be
working the queue it tests against — and CI installs `expect` rather
than accepting the named skip, because a skip is not a pass.

The tier earned its place on the first run. Bubble Tea reports height 0
to a terminal that has not sized itself, and both models floored that at
one line, so the entry screen rendered a single command above its
footer. Zero is now read as unknown and means no limit; a terminal that
*did* say, and said something too short to hold the chrome, still keeps
its line. Both models carry a unit case for the distinction.

**The screen is a session now, and that is a reversal.** This Outcome
argued the opposite above — that the screen must be gone before a
command asks anything, and that reading the next thing is another
`writrun`. That was the right conclusion from the wrong options: the
loop it rejected opened a *second* Bubble Tea program while the first
still held the keyboard, and `finish`'s confirmation answered itself
with the `enter` that had chosen `finish`. The fault was two programs,
not the returning.

There is one program now. The entry screen and the queue are its two
states, and a command runs through `tea.Exec`, which pauses it and
releases the terminal before the command asks anything. Two readers
never exist at once because there is only ever one. The maintainer
reported the closing three times — `work`, `take`, `config` — and it
was one behaviour, not three.

Two things this turned up. A command's output would be carried off by
the returning screen, so the reader says when they are done with it.
And the obvious way to wait for that — a buffered reader — reads ahead
and drops whatever it took past the newline, which is keys typed for
the screen: the same swallowing, one layer down. It reads one byte at a
time instead.

It is not provable at the model level, and the reason has the shape of
the bug: releasing the terminal means cancelling the read already in
flight, which a terminal allows and an `io.Reader` does not — a case
with a fake reader races itself. The pty tier proves it, and was held
against the old behaviour first: with the command ending the program it
fails naming that the screen never came back.

A command is a screen like the others, too: it takes the alternate
buffer rather than printing into the scrollback beneath the screen that
dispatched it, and gives the terminal back untouched. The cost is that
output taller than the window scrolls off with no scrollback to reach
it — `writrun <command>` on its own is where a long answer is read at
leisure, and a screen that pages its own output is the fix if that
bites.

`screens/README.md` carries the rules as they now stand, replacing the
one that said the screen does not come back.

**A question named no way out of itself.** `report` was reported as
having no way back, and it was not the session: huh spells its footer
from the field's own bindings and leaves the form's `ctrl+c` unnamed,
so every question read `enter submit` and stopped there — how to go
forward, never how to go back. `esc` now cancels alongside `ctrl+c`,
and the key is written where huh does render, one line under the
title.

Getting there cost a wrong answer worth recording. The first diagnosis
was that `waitForLine` accepted only `\n` while a raw terminal delivers
`\r`, which was plausible and false: the e2e passes either way, because
huh restores canonical mode before returning. The fix stayed — a return
key is a return key, and a unit case holds it — but it fixed nothing
the maintainer had seen. What settled it was opening a pty and reading
the frame, which said `enter submit` and nothing else.

That assertion lives in the pty tier and cannot live below it: huh
writes nothing to an output that is not a terminal, which is why every
case in `internal/term` reads the answer and never the drawing.

**`make ui` reported a refusal as a broken target.** Choosing `take`
with nothing available dispatches a command that exits 1 having already
explained itself, and make added `*** [ui] Error 1` underneath — which
was read as the door being broken rather than the queue being empty.
The target now lets exit 1 through and fails on anything else, so a
refusal reads as the answer it is and a crash still stops the run.

**The palette had no cases at all**, and its coverage came from the
screens that use it — so the promise its own doc makes, that a palette
told not to paint hands the text straight back, was held by nobody. It
has cases now, and writing them turned up the trap that makes a test of
this kind worthless: a test binary writes to a pipe, lipgloss sees no
terminal and renders plain, and the case passes with the guard deleted.
It names a colour profile so that it cannot. Both were held against a
deleted guard before being trusted.

**Two divergences the drawing carried and the plan did not name**, both
found by the maintainer rather than by a test. The screens ran inline,
so each left its last frame in the scrollback and the next drew beneath
it — they now take the alternate buffer, which is what the drawing's
full window always showed and what leaves a clean terminal for the
command a key dispatches. And the drawing's hues are a canvas's: the
binary paints with the terminal's indexed palette instead, so a value
chosen for one background is not imposed on another. `screens/README.md`
now states which of the two the drawing decides — where colour goes,
never which colour — because it was silent and the choice was made in
code alone.

The lesson is narrower than "check the drawing": layout, keys and
wording were checked against it, and *screen behaviour* was not.

**One test the plan asked for does not exist, and the file says why.**
Driving both screens through `Open` hangs: two Bubble Tea programs in
one process share one reader, and the first consumes what the second
would read. A live terminal blocks for the next key; a `strings.Reader`
does not. The transition is covered at the model level instead — `enter`
on `list` raises the queue, `esc` raises back, and neither dispatches.

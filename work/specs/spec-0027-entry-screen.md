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

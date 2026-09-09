---
id: spec-0034
task_ref: task-0031
status: implemented
created: 2026-09-09T06:24:00Z
---

# spec-0034 — `config` opens a screen, and a key is chosen on it

**References:** [task-0031](../tasks/task-0031-config-screen-navigation.md)

- **Goal:** `writrun config` with no argument opens the screen
  [`config.excalidraw`](../../docs/product/screens/config.excalidraw)
  draws, instead of printing the keys and returning.

## Scope

In: the screen — its rows, its movement, its detail line, its footer,
and `enter` calling the edit path that already exists.

Out: offering the allowed values before the write. That needs the
vocabulary the kit keeps to itself
([report-0030](../reports/report-0030-settings-vocabulary-unreadable.md)),
and this spec asks for a value the way the command already asks for
one — free text, judged by the checker after.

Also out: `writrun config <key>` and `writrun config <key> <value>`,
which keep their behaviour exactly. A script that calls them is not to
notice this change.

## Steps

1. Build the rows from the settings file the way `show` already does —
   the sections and their keys in the kit's order, values through the
   kit's reader. The screen holds no list of its own.
2. Make them a screen: a cursor on selectable rows only, movement that
   stops at the ends, a detail line naming the selected key and its
   value, and the footer the drawing gives.
3. On `enter`, ask for the new value and run the edit path
   [spec-0031](spec-0031-config-screen.md) built — write, check,
   restore on refusal.
4. Show the checker's verdict, its own sentence unedited, and return to
   the screen with the rows re-read.
5. `q` leaves. The screen is what `writrun config` opens, so leaving it
   ends the command.

## Acceptance criteria (EARS)

- When `writrun config` runs with no argument on a terminal in an
  adopted repository, the system shall open the screen rather than
  print the keys.
- When the screen opens, the system shall show every key the settings
  file names, under the section the kit gives it.
- When `↑` or `↓` is pressed, the system shall move between selectable
  keys and shall not wrap at either end.
- When a key is selected, the system shall show that key and its
  current value on the detail line.
- When `enter` is pressed, the system shall ask for a value and then
  run the existing write-then-check path, unchanged.
- When the checker refuses, the system shall print the checker's own
  message, leave the file byte-for-byte as it was, and return to the
  screen.
- When the checker accepts, the system shall return to the screen with
  the changed value shown.
- When `q` is pressed, the system shall leave, having changed nothing
  it had not already changed.
- When stdin or stdout is not a terminal, the system shall print the
  keys as it does today — a screen needs a terminal, and a script
  reading `writrun config` must keep reading it.

## Edge cases

- **The settings file is absent.** Unchanged: the command already
  refuses, naming that writing one is adoption's act.
- **A key whose value is long** — `commit_scopes` is a sentence. The
  row shows what fits and the detail line shows the whole value, as
  the entry screen already does with summaries.
- **The file changes underneath.** The rows are re-read after every
  accepted change rather than remembered, so the screen shows the file
  and not a copy of it.
- **A refusal.** The reader stays on the key they chose, so the
  checker's sentence and the key it is about are on the screen
  together.

## Tests required

- Unit, `internal/screen`: the rows come from the settings the caller
  passed and no list is held here; movement stops at both ends; the
  detail line names the selected key and its value.
- Unit, `internal/command/configcmd`: the no-argument path opens the
  screen where there is a terminal, and prints where there is not.
- Unit, `internal/command/configcmd`: an accepted change and a refused
  one behave as spec-0031 already requires — this spec adds a screen
  in front and moves none of that.
- e2e, `tests/e2e/screen/`: driven on a real pty — the screen draws its
  footer, the cursor moves, and `q` leaves. It changes no setting: a
  case that wrote one would be editing the repository it runs in.

## Definition of Done

- [ ] `writrun config` matches
      [`config.excalidraw`](../../docs/product/screens/config.excalidraw)
      in rows, movement and footer.
- [ ] No allowed value is written in Go; the checker remains the only
      authority on what a value may be.
- [ ] `writrun config <key>` and `writrun config <key> <value>` are
      byte-for-byte the same experience they are today.

## Proposed product changes

- `product/config.md` — says the no-argument form opens a screen, and
  that the two argued forms are unchanged.

## Outcome

Built as specified. `writrun config` with no argument opens the screen
the drawing gives — the keys under their sections, a cursor, a detail
line, and `↑↓ move · enter change · q back` — and `enter` runs the
write-then-check path spec-0031 already built. Without a terminal it
prints the listing, unchanged.

**The terminal handover is one thing now, not two.** A change asks a
question, and a question is another terminal program, so the settings
screen has to release the terminal exactly as the session does for a
command. Rather than write that discipline twice, `dispatch` became a
label and a `func`: the session hands over a command, this screen hands
over a change, and the part that is easy to get wrong lives in one
place.

**A defect this had, caught by a unit case and not by the pty.** The
rows carried the key's bare name and handed that back as its identity,
but the kit knows a key as `section.key` — so `enter` would have failed
to write *after* the reader had typed the value. The pty case missed it
because it cancels the question, which is what keeps it from editing
the repository it runs in. `Setting` now separates what is shown from
what identifies, and a case holds the two apart.

**A stream the frame did not carry.** A command that opens a screen
needs stdin, and `Ctx` had only the two writers — a question goes
through `Terminal`, which holds its own reader. `Ctx.Stdin` is nil
wherever the frame was built without one, which is every case that
asks nothing.

**What is still out of reach.** The screen offers no allowed values
before the write, because they are bash variables inside
`check_settings.sh`
([report-0030](../reports/report-0030-settings-vocabulary-unreadable.md),
open). A value is typed and judged after, on the screen as on the
command line. If that report is triaged, this is where the choices
would go.

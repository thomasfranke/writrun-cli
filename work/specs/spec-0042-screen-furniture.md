---
id: spec-0042
task_ref: task-0034
status: implemented
created: 2026-09-13T22:29:36Z
---

# spec-0042 — One header, one footer, and one way out on every screen

**References:** [task-0034](../tasks/task-0034-screen-furniture.md)

- **Goal:** Every screen opens with the same two lines, ends with the same one, and gives every key the same meaning it has on the screen before it.

## Scope

In: the header, the footer, the pager's own furniture, and the sentence
an unreadable read prints.

Out: what any screen lists — the rows are each screen's own.

## Steps

1. Give every screen a two-line header: the identity line
   `cmd/writrun/main.go`'s `header()` already composes, and a context
   line naming what this screen reads. The queue screen has neither
   today.
2. End every screen with one footer in three groups, in this order:
   movement, the actions one key each, the way out.
3. `esc` goes back and `q` quits. A screen with nowhere to go back to —
   the entry screen, the first run — offers `q quit` alone. The config
   screen's footer calls `q` a way back where the key quits, and names
   no quit at all; it is relabelled
   ([report-0035](../reports/report-0035-screen-furniture.md)).
4. Wrap a dispatched command's output in the same furniture: the pager
   is a screen. Run from a shell the command prints its own lines and no
   header, as [rules.md](../../docs/product/rules.md) requires of plain
   output.
5. Leave a question without header or `q` — `q` would be a value — and
   ending in `internal/term`'s `wayOut`, unchanged.
6. Make a failed read name the command that answers it, not only the
   error.

## Acceptance criteria (EARS)

- When any screen renders, it shall open with the identity line and a
  context line.
- When a screen has somewhere to go back to, its footer shall end
  `esc back · q quit`.
- When it has not, its footer shall end `q quit`, and shall not offer
  `esc`.
- When a command's output is shown in the pager, the pager shall carry
  the same header and the same way out.
- When a question is asked, it shall carry no header and no `q`.
- When a read fails, the message shall name the command that answers it.
- When a command runs from a shell, its output shall carry no header.

## Edge cases

- **A terminal narrower than the header.** The header truncates; it
  never wraps into the rows.
- **The settings screen with no settings file.** Its own state, drawn:
  it says so and offers no edit.
- **A screen with nothing selectable.** The footer drops the keys that
  cannot act and says `nothing to select`, as the queue already does.

## Tests required

Unit, `internal/screen`: each screen's first two lines and its last;
a table test asserting every footer ends in one of the two forms and
that no screen composes its own header.

Integration, `tests/integration/screen/`: the rendered frames match the
drawings, including the pager's.

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`entry.excalidraw`](../../docs/product/screens/entry.excalidraw)
  — `writrun — the entry screen`, `the furniture, on the queue screen as
  the worked example`, `the same two lines around a command's output`,
  `a command that asks nothing, while it runs`, `a command that asks,
  taking the terminal`, and `the queue could not be read`.
- [`tasks/list.excalidraw`](../../docs/product/screens/tasks/list.excalidraw)
  — `the queue screen — a task selected` and `the queue with nothing in
  it`.
- [`adoption/config.excalidraw`](../../docs/product/screens/adoption/config.excalidraw)
  — `writrun config` and `the settings could not be read`.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [x] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [x] Header and footer come from one place; no screen writes its own.
- [x] `q` quits on every screen, and `esc` is offered only where there
      is somewhere to go.
- [x] [report-0035](../reports/report-0035-screen-furniture.md) closed.

## Proposed product changes

- `product/screens/README.md` — the furniture every screen carries,
  stated as a rule.
- `product/screens/` — the drawings the rule is checked against. Two of
  them stated the one-liners the command table replaced, and one drew
  no `config` row for a command the entry screen lists
  ([report-0042](../reports/report-0042-stale-drawn-summaries.md)).

## Proposed technical changes

- `technical/layout/tree.md` — what `internal/screen/` holds, which is
  every screen rather than two of them, and `internal/drawing/` beside
  it.

## Outcome

Built. `internal/screen/furniture.go` composes the header and the
footer, and no screen composes either: a screen states its identity
line, its context line and the file it read, names its movement and its
actions, and says which of the three ways out it has. The entry screen,
the queue, the pager, the settings, the doctor screen, the first run,
the running state and the two unreadable states all go through it, and
`TestEveryScreenOpensWithTheSameTwoLines` and
`TestEveryFooterEndsInOneOfTheTwoForms` hold every one of them by name.

`q` quits everywhere and `esc` is offered only where there is somewhere
to go. The config screen's `q back` is gone: it ends `esc back · q
quit` like every other screen reached from another, and both keys
close it, which is what going back means for a screen that is its own
program.

The pager carries the same two lines and the same way out, and so does
the terminal a command that asks is handed — that one says whose
terminal it is and why, where it used to print two words. A failed read
is a screen now: the failure marked as `doctor` marks one, the message
under it, and the commands that answer it named.

Nine frames are asserted (`internal/screen/screens_frames_test.go`),
and `internal/screen/frames_test.go` lost its copy of the drawing
reader to `internal/drawing`, which grew a `Screen` reading beside
`Frame`: a transcript frame ends at its divider and a screen frame
draws its explanation under one, and that is the whole of their
difference.

[report-0035](../reports/report-0035-screen-furniture.md) is closed by
this, and
[report-0042](../reports/report-0042-stale-drawn-summaries.md) was
fixed on the way: `entry.excalidraw` and `first-run.excalidraw` stated
the one-liners the command table replaced, drew no `config` row, and
named the settings file at the address the two homes moved it from.

**What the plan did not foresee.**

The screens needed facts, not only lines. The entry screen's context
line names the stage and its subject with the settings file at the
right of it; the queue's counts the lister's own sections; the config
screen's names the file and the check that judges a change to it. Each
is a cheap read the caller already makes, and the header is where they
became visible.

Two of the entry drawing's frames are annotations of the screen rather
than states of it — the pager's and the two running frames' closing
paragraphs describe what the frame shows, and one of them names "the
two lines above and the one below", which is this spec's own step 4.
Those three are asserted by their header and their keys.

The worked example disagrees with `tasks/list.excalidraw` about the
same row: its context line counts a queue its rows do not show, and its
pane is a shorter sentence for the same task. It is captioned as the
furniture's worked example, so its header and its footer are what the
case holds — which is what it states.

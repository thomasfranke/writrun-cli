# Screens

**What `writrun` renders, drawn.** On any question of layout, keys or
wording these files are the reference the implementation is checked
against. The rules a drawing cannot carry are stated below them.

**A drawing decides where colour goes, never which colour.** Its hues
are a canvas's, chosen to read beside the other frames; the terminal's
are the reader's, and the binary paints with the indexed palette their
theme defines so a value chosen for one background is not imposed on
another ([rules](../rules.md)).

| | |
|---|---|
| [Queue](queue.excalidraw) | The lister's sections, navigated by keys. One keystroke in from the entry screen. |
| [Entry](entry.excalidraw) | What the entry screen lists, and how it reaches the rest. |
| [doctor](doctor.excalidraw) | [`doctor`](../adoption/doctor.md)'s report, grouped by stage, at every level a finding carries. |
| [Config](config.excalidraw) | [`config`](../config.md)'s keys, and what judges a change. |

## `writrun` with no command

Opens the queue as a screen navigated by keys; every action dispatches
a command. Which key does what is drawn in
[queue.excalidraw](queue.excalidraw).

- Requires a terminal on stdin and stdout; without one, prints what
  `--help` prints instead.
- Requires an adopted repository; outside one, prints what `--help`
  prints instead.
- Shows the sections [`list`](../queue/list.md) shows, in the same
  order.
- Dispatches [`take`](../pull-requests/take.md),
  [`work`](../queue/work.md) and [`status`](../queue/status.md), and
  nothing else.
- A key runs the command it names — its checks, its questions, its
  confirmation, unchanged ([rules](../rules.md)).
- **A command owns the keyboard alone, and the screen comes back when
  it is done.** A command asks through a terminal program of its own,
  and two of those in one process do not share a keyboard: a screen
  still holding the input is a question that answers itself. So the
  screen is paused and the terminal released — not closed — and reading
  the next thing is a keypress rather than another `writrun`.
- **A command's output is read before the screen returns.** The screen
  takes the whole terminal, so coming straight back would take the
  answer with it; the reader says when they are done with it.
- The screen offers no action a command does not already provide.
- The screen reads only; every change goes through the dispatched
  command.

## Rules for this folder

- A drawing answers how a screen looks; a rule answers what holds,
  including where no screen exists. Neither states the other's part.
- One file per screen. A screen drawn in two files is two answers about
  one screen.
- A frame is transcribed from a real run, or built from values read out
  of this repository's own files.
- A screen that does not exist yet is labelled proposed, in the table
  above and on the drawing.
- `.excalidraw` files open at [excalidraw.com](https://excalidraw.com)
  and in the VS Code extension.

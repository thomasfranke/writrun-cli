# Screens

**What `writrun` renders, drawn.** On any question of layout, keys or
wording these files are the reference the implementation is checked
against. The rules a drawing cannot carry are stated below them.

| | |
|---|---|
| [Queue](queue.excalidraw) | The lister's sections, navigated by keys. What `writrun` with no command opens today. |
| [Entry](entry.excalidraw) | What the entry screen lists, and how it reaches the rest. Proposed. |
| [doctor](doctor.excalidraw) | [`doctor`](../adoption/doctor.md)'s report, grouped by stage, at every level a finding carries. |
| [Config](config.excalidraw) | The ten keys of `.writrun/settings.json`. Proposed. |

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
- A key leaves the screen and runs the command it names — its checks,
  its questions, its confirmation, unchanged ([rules](../rules.md)).
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

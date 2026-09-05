# `writrun` with no command

Opens the queue as a screen navigated by keys; every action dispatches
a command.

- Requires a terminal on stdin and stdout; without one, prints what
  `--help` prints instead.
- Requires an adopted repository; outside one, prints what `--help`
  prints instead.
- Shows the sections [`list`](queue/list.md) shows, in the same order.

| Key | Does |
|---|---|
| ↑ ↓ | Moves the selection. |
| Enter | Runs [`take`](pull-requests/take.md) on the selected task. |
| `w` | Runs [`work`](queue/work.md) on the selected task. |
| `s` | Runs [`status`](queue/status.md). |
| `q` | Leaves the screen. |

- A key leaves the screen and runs the command it names — its checks,
  its questions, its confirmation, unchanged ([rules](rules.md)).
- The screen offers no action a command does not already provide.
- The screen reads only; every change goes through the dispatched
  command.

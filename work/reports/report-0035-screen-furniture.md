---
id: report-0035
status: tracked
task_ref: [task-0034]
doc_ref: product/screens/README.md
created: 2026-09-13T21:37:40Z
triaged: 2026-09-16T17:05:00Z
---

# the screens disagree on their header and their way out

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

Four screens render their own header and their own way out, and no two
agree.

`internal/screen/entry.go:232` ends `↑↓ move · enter run · q quit`;
`model.go:165`, the queue, ends `… esc back · q quit`; `pager.go:118`
ends `↑↓ scroll · esc back · q quit`; and `settings.go:273`, the config
screen, ends `↑↓ move · enter change · q back`.

The config screen's line is the one that lies. `settings.go:186` is
`case keyQuit, "ctrl+c", keyBack: return m, tea.Quit` — `q`, `esc` and
`ctrl+c` all quit, and the screen is a program of its own, so quitting
is a return to the shell. The footer calls that key `back`, offers no
`esc`, and names no quit at all. The unreadable state at
`settings.go:249` prints the same `q back`.

The headers diverge too, and further than a label. Only the entry
screen has one: `cmd/writrun/main.go`'s identity line, passed in as
`screen.Entry{Header: header(ctx)}`, with the stage line under it. The
queue screen opens with the lister's first row and the config screen
with its first settings row — neither names the binary, and neither
names what is being read.

A reader who learns `q` on the entry screen quits; one screen in the
same key quits too, under a word that says it will not.

**Triage — tracked.** [task-0034](../tasks/task-0034-screen-furniture.md)
carries it, through [spec-0042](../specs/spec-0042-screen-furniture.md):
two lines of header on every screen, one footer in three groups, `esc`
back and `q` quit. The drawings decide it rather than this report —
`product/screens/entry.excalidraw` holds a frame captioned *the
furniture every screen shares*, which says `esc` goes back and `q`
quits, never `q back`, and names this report where it says so.

It is not a relabel. `internal/screen/settings.go:186` collapses `q`,
`esc` and `ctrl+c` into one `tea.Quit`, and the drawn footer asks two of
them to differ — on a screen that is its own program and is reached two
ways, from the shell and from the entry screen. The words cannot be
corrected before the behaviour behind them is decided, and
`spec-0034`, which is `implemented`, is what put the current words
there.

Two corrections to the observation above. The config screen does have a
header — `internal/screen/settings.go:112` pushes one and
`internal/command/configcmd/configcmd.go:97` composes it — and what it
lacks is the identity line; the queue screen is the one with neither.
And there are seven footers rather than four: `model.go:163`,
`pager.go:120` and `session.go:443` draw three more, and the running
state draws none at all.

---
id: report-0035
status: open
task_ref: []
doc_ref: product/screens/README.md
created: 2026-09-13T21:37:40Z
triaged: null
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


---
id: report-0035
status: tracked
task_ref: [task-0034]
doc_ref: product/screens/README.md
created: 2026-09-13T21:37:40Z
triaged: 2026-09-13T22:20:00Z
---

# the screens disagree on their header and their way out

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

Four screens render their own header and their own way out, and no two
agree.

`internal/screen/entry.go:232` ends `↑↓ move · enter run · q quit`;
`model.go:165`, the queue, ends `… esc back · q quit`; `pager.go:118`
ends `↑↓ scroll · esc back · q quit`; and `settings.go:273`, the config
screen, ends `↑↓ move · enter change · q back` — the one screen where
`q` goes back rather than quitting, and the one with no way to quit at
all.

The headers diverge the same way. The entry screen and the config screen
open with `cmd/writrun/main.go`'s identity line and a context line under
it; the queue screen opens with the lister's first row and names neither
the binary nor what is being read.

A reader who learns `q` on the entry screen quits; the same key on the
next screen in takes them back instead.

**Triage — tracked.** [task-0034](../tasks/task-0034-screen-furniture.md)
carries it, through [spec-0042](../specs/spec-0042-screen-furniture.md):
one header, one footer, and `q` quitting on every screen.

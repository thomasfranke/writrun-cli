---
id: task-0034
status: backlog
blocked_reason: null
taken_by: null
spec_ref: [spec-0042]
doc_ref: product/screens/README.md
origin: rule
priority: high
depends_on: [task-0033]
milestone: null
created: 2026-09-13T22:29:19Z
queued: null
completed: null
merged: null
provenance: []
---

# Give every screen its furniture, and every plan a cursor

**References:** [product/screens/README.md](../../docs/product/screens/README.md) · [spec-0042](../specs/spec-0042-screen-furniture.md)

Four screens render their own header and their own way out, and no two
agree — `q` quits on the entry screen and goes back on the config
screen, which offers no way to quit at all
([report-0035](../reports/report-0035-screen-furniture.md)). Outside an
adoption there is no screen at all: the binary prints a list of thirteen
commands, twelve of which refuse to run there. And where there is a
screen, a plan of nine paths to remove or a list of task ids says
nothing about what any single line means.

[`screens/README.md`](../../docs/product/screens/README.md) and the
drawings under it now state one answer: two lines of header, one line of
footer in three groups, `esc` back and `q` quit; a screen where the kit
is absent; and a cursor with a footer on every plan and every question,
naming the row and the act the next key performs on it.

It matters most where the act is hard to reverse, and that is exactly
where the binary explains least.

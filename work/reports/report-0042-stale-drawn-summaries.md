---
id: report-0042
status: open
task_ref: []
doc_ref: product/screens/README.md
created: 2026-09-17T03:35:15Z
triaged: null
---

# two drawings state the one-liners the help drawing replaced

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

[spec-0039](../specs/spec-0039-commands-explain.md) rewrote the command
table's one-liners in the words
[`help.excalidraw`](../../docs/product/screens/help.excalidraw) draws,
and the entry screen shows the same strings — one field, two surfaces.
Two other drawings still state the strings that field held before.

`entry.excalidraw` draws eleven rows carrying the old wording, among
them `list        the queue: available, held back, untriaged`, where
the table now reads `see what work is waiting, and what is blocked`.
One of the eleven was never the table's: `doctor      examines stages
0–3, repairs nothing` appears at no commit reachable from `main`. The
same frame draws no `config` row, which the entry screen's own grouping
in `cmd/writrun/main.go` lists under adoption. `first-run.excalidraw` draws `init        install the
kit at v0.0.08 and ask the stage`, where the table reads `install
WritRun into this repository`.

The drawn row is also narrower than the strings now in the table. The
entry screen's window is 80 columns and the name field takes the first
14, so a summary has 66; the longest the table now holds is 51, and the
longest `entry.excalidraw` draws is 44. `cmd/writrun/main_test.go`
bounded summaries at 44 and named the drawing for it, so the bound
moved to the 66 the drawing's own window gives.

Two drawings of one field now state two different answers, and
`screens/README.md` says a drawing that must itself change is a
documentation change of its own, decided by a person.

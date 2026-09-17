---
id: report-0043
status: open
task_ref: []
doc_ref: technical/decisions/README.md
created: 2026-09-17T03:35:19Z
triaged: null
---

# a decision rejects the long descriptions a spec approved

**References:** [technical/decisions/README.md](../../docs/technical/decisions/README.md)

[decision 0010](../../docs/technical/decisions/docs/0010-help-is-one-line-per-command.md)
records `--help` as one line per command plus the docs' address, and
names full help text among the two answers it rejected: "a fork of the
product docs".

[spec-0039](../specs/spec-0039-commands-explain.md) approves the
opposite. Every command now carries a long description — a sentence and
four labelled parts — printed by `writrun <command> --help` and by a
bare run of a command that asks. Thirteen drawings under
`docs/product/screens/` state that text, and the spec's frames are what
the suite holds the binary to.

The decision is dated 2026-09-03 and carries no superseding note. The
spec promises `product/rules.md` and nothing under
`technical/decisions/`, so the implementing change could not touch it:
`writrun-check-spec-deltas` reads a doc the diff touches and no spec
promised as UNDECLARED.

A reader of `decisions/README.md` and a reader of `rules.md` now get
two answers about what `--help` prints.

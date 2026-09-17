---
id: report-0043
status: fixed
task_ref: []
doc_ref: technical/decisions/README.md
created: 2026-09-17T03:35:19Z
triaged: 2026-09-17T12:02:38Z
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

**Triage — fixed.**
[decision 0016](../../docs/technical/decisions/docs/0016-a-command-explains-itself.md)
records what is true now and says it supersedes 0010, which keeps its
file, its number and its text — `decisions/README.md` requires that an
entry is never edited and the one replacing it says so.

0016 does not overturn 0010 so much as answer what 0010 could not see.
Its objection was that help text written beside `product/` is a second
copy that drifts; what it lacked was a third place for the text to
live. The drawings are that place — every description is transcribed
from a frame, and the suite asserts the rendered lines against it, so a
copy that drifted fails the build instead of misleading a reader. The
cost 0010 named is still paid: the four commands run daily print no
description, and `--help` itself is unchanged.

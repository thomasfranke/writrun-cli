# the kit is read from the kit; Go names only what the binary calls.

**2026-09-06**

`internal/kitpaths` listed the kit's contents — three refresh
directories and four workflow files — and `doctorcmd` listed the four
human gates by their wording. WritRun `v0.0.04` added five files and
three gates, and every one of them needed a Go change to be seen; three
kit files the list never named had gone unrefreshed since adoption. The
inventory inverts: a refresh writes what the fetched template ships
minus the adopter-owned paths, and a check reads the rows of the file
the kit ships them in. What stays named in Go is what the binary calls —
the scripts, the settings, the tag, the pull-request template, the
queue's folders — and each of those is declared once. Ten of them were
declared in two to five packages each, so a script the kit moved was a
hunt rather than a line.

The exception is `AGENTS.md`, and it is one because the file is the
adopter's: no kit file can describe an edit to it, so
[`internal/pointer`](../../layout/tree.md) carries that shape alone. The
rule is [coupling](../../engineering/coupling.md).

Rejected: asking WritRun to ship a manifest — it would make the CLI's
convenience a constraint on the methodology, which
[about](../../../about.md) forbids ("optional client, never its
dependency"). Rejected: reading the kit's prose for its declarations —
the sentences are for agents, and a parser over them would break on a
rewording that changed nothing.

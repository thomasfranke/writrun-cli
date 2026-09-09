# Conventions

**These are WritRun's defaults, not your project's answers.** How
commits, branches, pull requests, tasks and specs are written — in
force while the matching file under `writrun/conventions/` defers to
it, and replaced the moment that file carries content of its own.

**Overriding is per file, and it is whole.** Replace
`writrun/conventions/commits.md` and your text stands entire; the six
beside it go on deferring, so a correction WritRun ships still reaches
them. Nothing merges in either direction, and nothing here is edited by
hand — the next update overwrites the edit, by design
([Two homes](https://github.com/thomasfranke/writrun/blob/main/docs/product/adoption.md#two-homes)).

WritRun's machinery reads none of this prose — mechanics depend only on
the grep-level markers listed in its
[public contract](https://github.com/thomasfranke/writrun/blob/main/docs/technical/distribution/README.md)
and on the values in `settings.json`. Agents read these before writing;
tooling (a commit-msg hook, a PR opener) validates against whatever
they say.

| | |
|---|---|
| [Commits](commits.md) | Conventional Commits — the subject's shape, and where the two vocabularies live |
| [Branches](branches.md) | Naming per flow, and the id marker the machinery reads |
| [Pull requests](prs.md) | Title rule, template, merge policy |
| [Tasks](tasks.md) | Title, body, priority and milestone taste |
| [Specs](specs.md) | Title, criteria, scope and Outcome taste |
| [Prose](prose.md) | How documentation, skills and comments are written |

One rule spans them all by default: **English everywhere** — code,
comments, commits, documentation. A project that writes in another
language says so in its own copy of this file.

Tooling needs these choices machine-readably — the scripts already act on
some of them — so the data lives in [`settings.json`](../../../writrun/settings.json),
at the root of the project's own home, and these `.md` files carry the reasoning:
what the options are and why a project would pick one. **Nothing is
stated in both.** A value here that also sits in the settings file is a
value that will eventually disagree with itself; if you find one, the
settings file wins and the prose is the bug.

**Three questions about the docs are answered there too, and never by
reading the file tree**: `stage_1.spec_required` says when a task needs
a spec, `stage_1.decisions_style` says where dated decisions live —
`per-subsystem` or one `chronological` log — and
`stage_1.product_layout` says how the product half is organized,
`by-concept` or `by-feature`. Each is a variant Adoption leaves open
and orders declared, so an agent asks the file rather than inferring a
shape from whichever folders it happened to open first.

That split was always the plan; the file is JSON rather than the
front-matter this once predicted, because it is edited by people who have
not read WritRun's front-matter contract and JSON is the shape they
already know. It no longer lives in this folder — the update exemption
`conventions/` carries moved onto the file by name when it took the root
address. See
[`decisions/0052`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/decisions/tasks-and-specs/0052-settings-carry-the-choice.md)
and [`decisions/0053`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/decisions/tasks-and-specs/0053-settings-at-the-root.md).

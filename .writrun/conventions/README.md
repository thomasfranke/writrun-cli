# Conventions

The adopter-editable layer: how this project writes commits, branches,
pull requests, tasks, and specs.

**Every file here is the project's to edit.** WritRun ships them as
defaults and its machinery reads none of them — mechanics depend only on
the grep-level markers listed in WritRun's
[public contract](https://github.com/thomasfranke/writrun/blob/main/docs/technical/README.md#distribution).
Agents read these before writing; tooling (a commit-msg hook, a PR
opener) validates against whatever they say.

| | |
|---|---|
| [Commits](commits.md) | Conventional Commits — types, scopes, examples |
| [Branches](branches.md) | Naming per flow, and the id marker the machinery reads |
| [Pull requests](prs.md) | Title rule, template, merge policy |
| [Tasks](tasks.md) | Title, body, priority and milestone taste |
| [Specs](specs.md) | Title, criteria, scope and Outcome taste |
| [Prose](prose.md) | How documentation, skills and comments are written |

One rule spans them all in this repository: **English everywhere** —
code, comments, commits, documentation.

Tooling needs these choices machine-readably — the scripts already act on
some of them — so the data lives in [`settings.json`](../settings.json),
at the root of WritRun's home, and these `.md` files carry the reasoning:
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

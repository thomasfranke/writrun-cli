---
id: report-0023
status: authored
task_ref: []
doc_ref: product/pull-requests/amend.md
created: 2026-09-06T01:30:21Z
triaged: 2026-09-06T04:24:32Z
---

# amend.md says neither what the command writes nor that its pull request opens ready

**References:** [product/pull-requests/amend.md](../../docs/product/pull-requests/amend.md)

`docs/product/pull-requests/amend.md` predates the command — it landed
in the docs pass (#2) and says the shape holds and that the pull request
"says why". It does not say what `writrun amend` writes, and it does not
say that the pull request opens **ready for review** rather than as a
draft. Its three siblings do say both kinds of thing: `finish.md` — "What
it writes, after its checks pass and never before, is the spec's
`implemented` and the task's `completed` date"; `take.md` — the flow
"ends in a draft pull request", plus the `--title` question. Opening
ready is the one thing amend does differently from all three, and
`status: draft` on a spec is the one queue edit it makes; neither
appears anywhere under `docs/`. spec-0011 promised no product change and
`check_deltas.sh` enforces that promise, so this was seen while
implementing it and not acted on.

**Triage — authored.** No rule said what `amend` writes or how its pull
request opens, and `amend.md` now says both, plus that the body names
the pull request the amendment suspends.

**A correction to this report.** It says opening ready is the one thing
`amend` does differently from all three siblings. `author` opens ready
too — its package doc says "opened **ready**, never draft" and its
suite asserts `--draft` is absent — and `author.md` was equally silent.
The chapter was short one sentence on two pages, not one, so
`author.md` gained the same line in the same change.

Derivation is none: the code already does what the new sentences say.
---
id: report-0023
status: open
task_ref: []
doc_ref: product/pull-requests/amend.md
created: 2026-09-06T01:30:21Z
triaged: null
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

---
id: report-0019
status: open
task_ref: []
doc_ref: product/pull-requests/finish.md
created: 2026-09-06T01:30:24Z
triaged: null
---

# finish.md says the writes happen after the checks pass and never before, which the undo does not make true

**References:** [product/pull-requests/finish.md](../../docs/product/pull-requests/finish.md)

`docs/product/pull-requests/finish.md` says of `writrun finish`: "What
it writes, after its checks pass and never before, is the spec's
`implemented` and the task's `completed` date." The command writes both
at step 2, before step 4 runs `preflight.sh` — that ordering is the
premise spec-0017 examined and kept, since preflight's completion
warning reads the `completed` date off the working tree. spec-0017
weighed `shape.md`'s "a refused command leaves nothing behind" against
spec-0010's step order and answered with an undo that makes the end
state right; it did not survey `finish.md`, whose sentence is about
*when* the writes happen and is not repaired by putting them back
afterwards. The spec's Proposed product changes say "none — `shape.md`
already states the rule", and the pull request implementing it states
"No `docs/` change", so nothing in that change looked at this line.
`shape.md`'s own "no status is written" bullet reads as an end-state
claim and survives; this one does not.

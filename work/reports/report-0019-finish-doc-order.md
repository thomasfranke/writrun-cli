---
id: report-0019
status: fixed
task_ref: []
doc_ref: product/pull-requests/finish.md
created: 2026-09-06T01:30:24Z
triaged: 2026-09-06T04:24:32Z
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

**Triage — fixed.** The sentence moved, not the code. `finish.md` now
says what the command writes and, separately, that the writes precede
the last of its checks.

The order is load-bearing, and the load sits in a script this
repository does not own. `preflight.sh` builds a warning per task whose
`completed` is null — "this run precedes the completion edits and does
not stand for them; run it again after them" — reading the working tree
on both its pass and its fail path. Writing after preflight would print
that warning on every `writrun finish` run, which is a worse output
than the wrong adjective was.

spec-0017 had already ruled, and named this report as the repair it
promised not to make. Reversing the order now would mean amending two
implemented specs; `writrun amend` refuses one — "spec-0017 is
'implemented' … an implemented one is history" — so the code route
costs a report branch, a task, a spec and an approval, to buy back one
adjective and lose preflight's clean verdict.

`shape.md`'s opening sentence survives narrowly, on its own "in the
order the methodology fixed" deferral. That it lists an order neither
`finish` nor `amend` literally follows is [[report-0026]].
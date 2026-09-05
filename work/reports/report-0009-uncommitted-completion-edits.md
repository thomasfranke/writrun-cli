---
id: report-0009
status: open
task_ref: []
doc_ref: product/pull-requests/finish.md
created: 2026-09-05T18:21:56Z
triaged: null
---

# The completion edits finish writes are never committed, so preflight judges a range without them

**References:** [product/pull-requests/finish.md](../../docs/product/pull-requests/finish.md)

`writrun finish` writes the spec's `implemented` and the task's
`completed` into the working tree and never commits them, while step 4's
`preflight.sh` reads its two later stages against a commit range —
`check_promised_deltas.sh` derives "the specs this change implements"
from the front matter at the range's two ends, and `check_state.sh`
reads its transitions the same way. On the fixture the sequence
therefore ends `PREFLIGHT OK — deltas checked: none — no spec reached
'implemented' in this range`: the two stages that exist to judge the
completion edits pass having seen none of them, and only stage 1's
whole-queue sweep read anything the command had just written. The order
is spec-0010's own and the writes cannot move after preflight — the
completion warning preflight prints is read off the working-tree
`completed` date — so what the gates vouch for at the moment `writrun
finish` runs is the branch as it was before the command touched it.

---
id: task-0037
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0045]
doc_ref: product/pull-requests/finish.md
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-09-17T13:05:44Z
queued: 2026-09-17T13:13:57Z
completed: null
merged: null
provenance: []
---

# Make finish's gates read the tree they vouch for

**References:** [product/pull-requests/finish.md](../../docs/product/pull-requests/finish.md) · [spec-0045](../specs/spec-0045-gates-read-the-tree.md)

`writrun finish` writes the spec's `implemented` and the task's
`completed` date into the working tree and commits nothing. It then
hands its checks the range `origin/main...HEAD` — the branch as last
committed — so the two stages that exist to judge those very edits run
without them. The sequence ends `PREFLIGHT OK — deltas checked: none —
no spec reached 'implemented' in this range`, and what the gates vouched
for is the branch as it was before the command touched it.

[report-0014](../reports/report-0014-uncommitted-completion-edits.md)
recorded it, and [spec-0017](../specs/spec-0017-finish-write-order.md)'s
Outcome says in writing that moving the writes cannot reach it — only
the range can.

Give the checks a range that reaches the working tree, so the edits
`finish` has just made are inside what it judges.

It matters because of the shape of the failure. A gate that refuses is
read and answered; a gate that passes over a change it never saw prints
a sentence nobody knows is empty — and this one prints `OK` under the
word `PREFLIGHT`, at the last moment before the forge is asked for
anything. Every completion since the command shipped has been vouched
for that way, including the three this repository merged this week.

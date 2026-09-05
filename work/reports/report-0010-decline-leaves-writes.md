---
id: report-0010
status: open
task_ref: []
doc_ref: product/pull-requests/shape.md
created: 2026-09-05T18:22:16Z
triaged: null
---

# A declined finish leaves the spec implemented and the completion date behind

**References:** [product/pull-requests/shape.md](../../docs/product/pull-requests/shape.md)

`shape.md` says a refused pull-request command leaves nothing behind —
"no half-written status, no orphan branch" — while spec-0010 fixes an
order in which the two writes are step 2 and the confirmation is step 5,
because step 4's `preflight.sh` has to read the `completed` date the
writes put there. Saying no to `writrun finish` therefore exits 1 with
the spec already `implemented` and the task already carrying its
completion date; the integration case
`a_decline_reaches_nothing_test.sh` asserts only that the forge stayed
untouched, which is the half of the sentence the implementation can
keep. The same shape is what `writrun author` and `writrun amend` will
inherit, so whichever of the two texts is wrong is worth settling before
they are written rather than after.

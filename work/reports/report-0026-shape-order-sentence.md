---
id: report-0026
status: open
task_ref: []
doc_ref: product/pull-requests/shape.md
created: 2026-09-06T04:25:03Z
triaged: null
---

# shape.md lists an order that neither finish nor amend literally follows

**References:** [product/pull-requests/shape.md](../../docs/product/pull-requests/shape.md)

`product/pull-requests/shape.md` opens by giving `author`, `take`,
`finish` and `amend` one shape: "run the repository's checks in the
order the methodology fixed, write the status the flow calls for,
assemble the branch, the commit title, and the pull-request body, show
all of it, and open the pull request on confirmation." Two of the four
do not follow that order literally. `finish` writes at step 2 and runs
`preflight.sh` after — the timing report-0019 corrected on
`finish.md` — and `amend` computes its queue edit at step 2 but writes
it beside the push, after the confirmation, which `amendcmd`'s package
comment states and which spec-0011's Outcome records as required rather
than optional. The sentence survives on its own "in the order the
methodology fixed" deferral, which declines to fix the internal order
and so cannot be contradicted by one. Whether a sentence that four
commands are measured against should be that elastic is what this
records; report-0019 repaired the page that carried no such deferral,
and left this one standing.

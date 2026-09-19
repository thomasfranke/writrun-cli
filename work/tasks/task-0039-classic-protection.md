---
id: task-0039
status: done
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0047]
doc_ref: product/adoption/doctor.md
origin: report
priority: high
depends_on: []
milestone: null
created: 2026-09-19T21:41:29Z
queued: 2026-09-19T22:17:43Z
completed: 2026-09-19T22:39:42Z
merged: 2026-09-19T23:23:26Z
provenance: []
---

# doctor reads every rule that governs main, not rulesets alone

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md) · [spec-0047](../specs/spec-0047-classic-protection.md)

Stage 2 asks the forge two questions about `main`: whether anything
governs it, and whether a rule over it refuses the recording push. Both
answers come from one read, `rules/branches/main`, which reports the
rules rulesets contribute. A classic branch protection rule is not one
of them, so a branch a classic rule governs is examined as if nothing
governed it.

Make both rows read every rule the forge holds over `main`, the classic
rule included, and name what they read: the row, the remedy sentence and
the table in `product/adoption/doctor.md` say `ruleset` where they now
mean any rule over the branch.

It matters because the blind spot inverts both answers at once. On the
repository [report-0048](../reports/report-0048-doctor-reads-only.md)
records, the classic rule required a pull request and enforced it on
admins, on a user-owned repository where the forge offers the Actions
bot no bypass actor. doctor printed `✓ no rule over main refuses the
recording push` and graded the governance it could not see as a
recommendation. The recording push would have been refused on the first
merge — the fault stage 2 exists to find before a project declares that
stage.

The governance row also states and denies one fact in one sentence:
`main is governed by a ruleset — no ruleset governs it — the methodology
recommends protecting it`. The requirement's name and this run's answer
are rendered with no seam between them, and this is the only row that
reaches that state.

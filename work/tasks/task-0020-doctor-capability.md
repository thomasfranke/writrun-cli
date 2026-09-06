---
id: task-0020
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0019]
doc_ref: product/adoption/doctor.md
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-06T00:37:31Z
queued: 2026-09-06T00:53:10Z
completed: 2026-09-06T04:04:26Z
merged: null
provenance: []
---

# Have doctor check that the recording push can land, not how the forge is configured

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md) · [spec-0019](../specs/spec-0019-doctor-capability.md)

`doctor.md` requires Actions workflow permissions set to read-and-write
and the Actions bot on the ruleset's bypass list. This repository has
neither and its recording pushes land anyway: workflows raise their own
`contents: write`, and no rule the ruleset enables refuses a
fast-forward push. So `writrun doctor` reports two findings against a
repository that is configured correctly — and more tightly than the
document asks for.

The stage-2 check reads how the forge is configured where it should read
whether the recording push can land. Two adopters can both be correct
and configured differently; a check that names one configuration fails
the other for no defect. See [[report-0013]], declined for this reason.

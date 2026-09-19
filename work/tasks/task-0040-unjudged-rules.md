---
id: task-0040
status: backlog
blocked_reason: null
taken_by: null
spec_ref: [spec-0048]
doc_ref: product/adoption/doctor.md
origin: report
priority: high
depends_on: []
milestone: null
created: 2026-09-19T23:33:54Z
queued: null
completed: null
merged: null
provenance: []
---

# A rule this binary does not judge is named, never passed in silence

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md) · [spec-0048](../specs/spec-0048-unjudged-rules.md)

Stage 2 keeps a hand-written list of four rules that refuse the
recording push and reports every other rule over `main` as refusing
nothing. The forge's own vocabulary is twenty-five ruleset rule types.
Ten of the twenty-one doctor does not name can refuse a fast-forward
push by the Actions bot — `required_deployments`, `workflows`,
`code_scanning`, the three commit message and email patterns, and the
four file restrictions — and a branch protection rule carries
`lock_branch`, which makes the branch read-only outright
([report-0051](../reports/report-0051-locked-branch.md)).

Invert the default. A rule this binary has not judged is named as one it
has not judged; only the rules a plain fast-forward push demonstrably
meets pass in silence. `lock_branch` is judged: the forge documents it
as the branch nobody can push to.

It matters because the reader cannot tell the two apart today. `✓ no
rule over main refuses the recording push` is printed whether doctor
examined the rule and cleared it or never knew the rule existed, and the
methodology's own gate for declaring stage 2 is that line. A vocabulary
that ages every time the forge adds a rule type is a check that quietly
stops checking.

The grade a named-but-unjudged rule carries, and whether it is a row of
its own, are the spec's to settle: the row that exists today asserts
that nothing refuses the push, and a rule nobody judged cannot sit
under a sentence saying that.

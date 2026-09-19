---
id: report-0051
status: tracked
task_ref: [task-0040]
doc_ref: product/adoption/doctor.md
created: 2026-09-19T22:38:47Z
triaged: 2026-09-19T23:33:54Z
---

# A locked branch refuses every push, and the four blocking rules do not name it

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md) · [task-0040](../tasks/task-0040-unjudged-rules.md)

`blockers` in `internal/command/doctorcmd/forge.go` names four rules
that refuse the recording push: restrict updates, require signed
commits, require status checks to pass, and require a pull request
before merging. A branch protection rule carries a fifth field the same
sentence is true of. `lock_branch` makes the branch read-only, and the
documented response schema of `GET /repos/{owner}/{repo}/branches/{branch}/protection`
carries it beside the four.

A `main` that is locked therefore reports `✓ no rule over main refuses
the recording push`, and the push is refused on the first merge — the
same shape [report-0048](report-0048-doctor-reads-only.md) recorded for
a classic rule as a whole, one field further in.

spec-0047 put the vocabulary out of scope in as many words: "Out: the
four rules `blockers` names, which this spec maps onto rather than
grows". The classic rule's other fields are mapped onto those four and
`lock_branch` is not, so task-0039 leaves this standing by its own
brief. Nothing was reproduced against a locked branch: the gap was read
out of the API's schema while the mapping was written.

# Specs — the detail of one change

**Historical record — not a description of the present.** One file per
spec, named `spec-NNNN-<subject>.md` — the same shape a task's name has,
four-digit id and subject slug. A spec belongs to exactly one task
(`task_ref`) and inherits order and priority from it. Schema in WritRun's
technical README (`docs/technical/README.md#spec-schema` in the WritRun
repository).

Lifecycle: `draft → approved → implemented`. Approval is a human gate —
an agent never self-approves. Content under an approval never changes
silently: an amendment returns the spec to `draft` and passes the gate
again. An implemented spec is never edited beyond its **Outcome** section
— divergence is documented, not erased.

The two **Proposed changes** sections are the merge contract: the
completing diff must touch everything listed and nothing permanent that
isn't.

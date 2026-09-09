---
id: task-0030
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0033]
doc_ref: technical/engineering/coupling.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-08T16:15:26Z
queued: 2026-09-08T20:14:58Z
completed: null
merged: null
provenance: []
---

# Follow WritRun's two homes: the adopter's files leave .writrun/

**References:** [technical/engineering/coupling.md](../../docs/technical/engineering/coupling.md) · [spec-0033](../specs/spec-0033-two-homes-migration.md)

WritRun `v0.0.05` moved the adopter's files out of the folder an update
replaces, and `v0.0.07` layered them: `.writrun/` is the kit's whole and
now carries `defaults/`, while `writrun/` is the project's, one file per
answer that either defers to a default or overrides it whole. The commit
vocabulary became two settings the checks read. This binary pins
`v0.0.04` and still names the old addresses.

Until it follows, the pin cannot move. `writrun update` run against
`v0.0.07` today would write the kit's shipped defaults to
`writrun/settings.json` — the address the new kit scripts read — and
leave this project's real answers stranded at `.writrun/settings.json`,
which nothing reads any more. The project would silently become stage 1
with `agent_coauthor: true`, a value its own `AGENTS.md` forbids.

The behaviour is not this repository's to invent. WritRun's spec-0090
wrote the contract and put the client out of its own scope: the update
moves the three adopter files once, then never touches `writrun/`
again; where both addresses exist the new one wins and the old one is
reported, never silently merged.

It matters because every adopter already on `v0.0.04` meets this the
first time they update, and the failure is silent. It matters twice
because the shape this leaves behind decides what the next tag costs:
the aim is a `v0.0.08` that is a constant to change, not a migration to
design.

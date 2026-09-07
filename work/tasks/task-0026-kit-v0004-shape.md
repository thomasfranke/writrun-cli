---
id: task-0026
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0025]
doc_ref: technical/engineering/coupling.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-06T20:34:59Z
queued: 2026-09-06T21:49:00Z
completed: 2026-09-07T00:03:26Z
merged: null
provenance: []
---

# Refresh and adopt the kit shape WritRun v0.0.04 ships

**References:** [technical/engineering/coupling.md](../../docs/technical/engineering/coupling.md) · [spec-0025](../specs/spec-0025-kit-v0004-shape.md)

This binary pins WritRun `v0.0.03`. At `v0.0.04` the kit keeps its agent
flow in a file of its own and the project's gate answers in another, and
carries three files this binary's inventory does not name. `writrun
update` reads the fenced section that shape no longer has, so it stops
and refreshes nothing: the pin cannot move until the four adoption
commands read the new shape.

Bring `init`, `update`, `uninstall` and `doctor` under
[coupling](../../docs/technical/engineering/coupling.md), move the pin
to `v0.0.04`, and refresh this repository's own kit with the result —
the binary adopts the methodology it packages, so a refresh it cannot
perform here is one no adopter can perform either.

Teaching the four commands this one tag by hand would leave the next
tag exactly where this one found them. The rule is what the work is
against: the inventory comes from the fetched template, the gates from
the rows of the file that states them, and `AGENTS.md` stays the single
shape the binary knows.

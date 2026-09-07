---
id: task-0028
status: backlog
blocked_reason: null
taken_by: null
spec_ref: [spec-0027]
doc_ref: technical/engineering/coupling.md
origin: rule
priority: medium
depends_on: [task-0027]
milestone: null
created: 2026-09-07T03:15:39Z
queued: null
completed: null
merged: null
provenance: []
---

# Separate the kit from the adopter at the declared boundary

**References:** [technical/engineering/coupling.md](../../docs/technical/engineering/coupling.md) · [spec-0027](../specs/spec-0027-adopter-boundary.md)

`writrun uninstall` deletes `.writrun/` whole. The adopter's
`settings.json`, `gates.md` and `conventions/` die with it — the very
paths `internal/kitpaths` declares `Untouchable`. The kit-shipped
`CLAUDE.md` shim survives the same uninstall as residue. `init` and
`doctor` require `docs/about.md`, a `docs/product/` chapter and a
`docs/technical/` doc; the kit's own instructions call that shape
optional. No adopter-owned text states the boundary in one place for an
agent working in an adopted repository.

Make removal keep every adopter answer and leave no kit residue. Hold
the binary to enforcing only what the kit declares. State the boundary
contract in `AGENTS.md`: what the adopter edits, what is never edited
by hand, and that a kit defect is a local report routed upstream.

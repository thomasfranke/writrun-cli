---
id: task-0027
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0026]
doc_ref: technical/engineering/coupling.md
origin: rule
priority: medium
depends_on: [task-0026]
milestone: null
created: 2026-09-06T21:22:09Z
queued: 2026-09-06T21:49:00Z
completed: 2026-09-07T00:35:37Z
merged: null
provenance: []
---

# Declare each kit path once, in the package that owns the act

**References:** [technical/engineering/coupling.md](../../docs/technical/engineering/coupling.md) · [spec-0026](../specs/spec-0026-kit-paths-once.md)

Ten kit paths are declared in two to five packages each. `list_tasks.sh`
is a constant in four files under four names — `listScript`,
`listScript`, `listerScript`, `lister`. `work/tasks` is a constant in
six. `read_setting.sh` and `check_observance.sh` are in three apiece,
`preflight.sh`, `check_front_matter.sh` and
`pull_request_template.md` in two.

A path the kit moves is a hunt through the tree rather than one line,
and a rename that misses a copy compiles. Declare each path once, in the
package that owns the act, and have every command reference that name.

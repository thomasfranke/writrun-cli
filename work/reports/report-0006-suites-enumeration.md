---
id: report-0006
status: open
task_ref: []
doc_ref: technical/testing/suites.md
created: 2026-09-05T16:58:56Z
triaged: null
---

# suites.md names two bash suites where the tree holds six

**References:** [technical/testing/suites.md](../../docs/technical/testing/suites.md)

`suites.md` names the bash suites as two: the release suite under
`tests/*/release/` and the CLI cases under `tests/integration/cli/`.
The tree holds six suite directories under `tests/integration/` alone —
`cli`, `coverage`, `init`, `release`, `uninstall`, `update` — all on
the same `tests/harness.sh`. The newest, `tests/integration/coverage/`,
arrived with task-0015; `init`, `uninstall` and `update` predate it, so
the enumeration was already short before that merge.

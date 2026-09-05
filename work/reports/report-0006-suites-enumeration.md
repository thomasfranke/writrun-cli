---
id: report-0006
status: fixed
task_ref: []
doc_ref: technical/testing/suites.md
created: 2026-09-05T16:58:56Z
triaged: 2026-09-05T17:19:00Z
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

**Fixed here, by removing the enumeration rather than extending it.**
Nothing else in the repository lists the suites: `tests/run.sh` walks
the tree, and `make test-<suite>` resolves by glob. A hand-kept list was
a fork of that, and it had already drifted twice — it also missed
`tests/e2e/adopt/` and `tests/e2e/release/`, so the count was eight, not
six. `suites.md` now states the discovery rule and tables the fixtures,
which is the fact it alone carries.

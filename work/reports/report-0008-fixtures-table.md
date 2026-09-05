---
id: report-0008
status: open
task_ref: []
doc_ref: technical/testing/suites.md
created: 2026-09-05T18:20:02Z
triaged: null
---

# The fixtures table names five of the eight fixtures

**References:** [technical/testing/suites.md](../../docs/technical/testing/suites.md)

`docs/technical/testing/suites.md` tables the fixtures and names what
sources each: `harness.sh`, `cli_lib.sh`, `coverage_lib.sh`,
`release_lib.sh` and `init_lib.sh`. The tree holds three more —
`tests/list_lib.sh` and `tests/take_lib.sh` landed with task-0006 and
task-0009, and `tests/report_lib.sh` arrives with task-0008 — so the
table names five of eight, and the three suites they serve
(`tests/integration/list/`, `/take/`, `/report/`) appear in no row.
report-0006 removed the suite *enumeration* from this same file because
a hand-kept list had already drifted twice; the fixtures table beside it
was kept, and it has drifted the same way since. Writing anything under
`docs/` is Thomas's gate in AGENTS.md, and task-0008 carries no spec whose
Proposed changes would authorise the edit.

---
id: report-0032
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-09T16:00:03Z
triaged: 2026-09-09T16:00:15Z
---

# A case needing expect had no guard, and release readiness has no expect

`tests/integration/suite/cases_never_read_the_operators_terminal_test.sh`
stands a terminal up through `expect` and had no guard for `expect`
being absent. `tests.yml` installs it; `release-readiness.yml` does not,
so every push to main since #123 has been red with
`expect: command not found` — twice per run, once directly and once
inside the release rehearsal.

The pty tier already had the guard and the idiom is written down. I
wrote the new case, in the same change, without it.

Two fixes, because a skip alone would leave main verified more weakly
than a pull request: the case now skips by name where `expect` is
absent, and `release-readiness.yml` installs `expect` so the case
actually runs there.

Seen on main at 9c335ba, and observed by the maintainer as "a pipeline
quebrou" before I had noticed it — the workflow only runs post-merge,
so no pull request check catches it.

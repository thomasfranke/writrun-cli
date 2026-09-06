---
id: report-0011
status: fixed
task_ref: []
doc_ref: technical/testing/suites.md
created: 2026-09-05T18:21:04Z
triaged: 2026-09-05T19:03:11Z
---

# The fixtures table in suites.md lists five of the eight fixtures

**References:** [technical/testing/suites.md](../../docs/technical/testing/suites.md)

`tests/` holds eight fixtures — `harness.sh`, `cli_lib.sh`,
`coverage_lib.sh`, `init_lib.sh`, `list_lib.sh`, `release_lib.sh`,
`take_lib.sh` and, from task-0014, `status_lib.sh`. The **Fixtures**
table in `suites.md` names five of them: `list_lib.sh` and `take_lib.sh`
were never added, and `status_lib.sh` is not added by this change
because spec-0013 proposes no doc change and a permanent doc the spec
did not promise may not be touched. The section above the table states
that no file lists the suites and that a case is registered by its path
— which is true of the cases and not of this table, so the one list
that has to be maintained by hand is the one nothing checks.

**Fixed here, with [[report-0009]], which is this same finding seen
independently on another branch.** This report named the reason the
other did not: the paragraph above the table claimed nothing lists the
suites while the table below it was a hand-kept list nothing checked.
The table is gone rather than corrected.

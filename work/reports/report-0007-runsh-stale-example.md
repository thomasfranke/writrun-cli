---
id: report-0007
status: fixed
task_ref: []
doc_ref: technical/testing/suites.md
created: 2026-09-05T17:18:38Z
triaged: 2026-09-05T19:03:11Z
---

# run.sh names a case file that does not exist

**References:** [technical/testing/suites.md](../../docs/technical/testing/suites.md)

`tests/run.sh`'s header comment offers a standalone invocation as its
example: `bash tests/integration/release/minor_bumps_third_digit_test.sh`.
No such file exists. The release suite's case is
`minor_bumps_middle_digit_test.sh`, and the middle digit is what a minor
bump moves — so the name in the comment is wrong twice, in path and in
meaning. Found while rewriting `suites.md`, which had copied the same
example; the doc now names the file that exists.

**Fixed here.** `tests/run.sh`'s header now names
`minor_bumps_middle_digit_test.sh`, which exists. Two further errors in
the same comment went with it: it described a `unit/` tier directory
that `tests/` does not hold — the unit tier is Go beside the code
(`tiers.md`) — and it named `tests/release_lib.sh` as though one fixture
served every case.

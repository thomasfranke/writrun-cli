---
id: report-0031
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-09T15:32:20Z
triaged: 2026-09-09T15:32:30Z
---

# The suite handed every case the operator's terminal

`tests/run.sh` and the Makefile's two case loops ran each case
with the stdin they were given, so a suite run from a keyboard handed
that keyboard to every case.

Most cases do not care. The ones that check what a command does
*without* a terminal do: they assert about the suite's own stdin. Given
a real one, `writrun report` asked its confirmation for real and waited.

Where that bites is `make release`. The release runs the suite; the
`e2e/release` case copies the tree and runs a whole release inside a
clone, capturing its output to assert on it; the nested suite inherits
the operator's terminal; and `the_question_precedes_the_write_test.sh`
asks its question inside that captured output. Nobody sees it, nobody
answers it, and the release neither ends nor says why. Observed twice
on 2026-09-09, the second time for over half an hour at 9b507f3.

It also explains a whole session of `make tests` runs that died at
`e2e/release` and were read as slowness or as an environment timeout.
They were hanging here. The one run that completed had been submitted
to launchd, which gives /dev/null as stdin — so its cases took the
no-terminal path and passed. The evidence was there and was read wrong.

Fixed by giving every case `</dev/null` in `tests/run.sh` and in both
Makefile loops: a case that wants a terminal stands one up itself
through WRITRUN_TTY_IN, and one that does not want it is now asserting
about a stdin the suite owns rather than about whoever ran it.

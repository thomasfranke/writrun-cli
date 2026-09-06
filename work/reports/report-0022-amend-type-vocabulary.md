---
id: report-0022
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T01:30:12Z
triaged: null
---

# writrun amend judges no branch-prefix vocabulary and the forge is where that lands

`writrun amend --type` accepts any lowercase word:
`internal/command/amendcmd/amendcmd.go`'s `plainWord` now requires a
letter and refuses a leading or trailing dash, so `--type -` no longer
composes the branch `-/slug`, but `--type wibble` still composes the
branch `wibble/amend-command`, the title `[Wibble][Specs] …` and the
commit subject `wibble(specs): …`. `check_observance.sh` faults all
three against `TYPES="docs feat fix refactor chore"`, and it does so on
the forge, after the branch is pushed and the pull request is open — the
one state `act`'s own error text calls the state this command must not
leave behind. Validating locally was not done because
`docs/product/rules.md` says no command reimplements a check and
`check_observance.sh`'s own comment says its list "is the machine half
of the same statement, and the two are kept in step by hand", so a third
copy in Go would be a third authority; `conventions/commits.md` is prose
rather than a format the binary can parse. `writrun take` does not have
the problem: it delegates to `take_task.sh`, which runs `valid_summary`
before anything is composed.

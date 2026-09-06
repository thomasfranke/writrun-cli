---
id: report-0021
status: declined
task_ref: []
doc_ref: null
created: 2026-09-06T01:29:55Z
triaged: 2026-09-06T04:26:48Z
---

# The pull-request body composer is written twice, once in Go and once in the kit

`internal/command/amendcmd/compose.go` composes the pull-request body by
stripping the adopter's template: `strip`, `replaceImplements`, `fill`,
`squeeze` and `sameLines`, about 130 lines, plus a `fallbackBody`
constant for a repository whose template is gone.
`.writrun/scripts/stage-2-pull-requests/take_task.sh` already does the
same three things in a ten-line awk program at lines 227-260 — drop the
guidance comments, drop the `## Derived work` half, replace the
`Implements spec-NNNN.` line — and its own fallback body is the same
text the Go constant carries, one heading apart. The duplication pull
request #62 declared was with `finishcmd`, a sibling command; this one
is with the methodology's own script, which `docs/product/rules.md`
names as the authority a command packages rather than reimplements. The
two will drift when the template changes, and only one of them is
covered by the kit's own suite.

**Triage — declined.** The rule this cites reaches checks, not
compositions. `product/rules.md` says "No command reimplements a
**check**"; `pull-requests/shape.md` gives the four commands one shape
that includes assembling "the branch, the commit title, and the
pull-request body". Composing a body is the documented job.

There is also nothing to reimplement. The kit has no amendment script,
and its one composer — `take_task.sh` — composes a task-taking pull
request, which `take` calls rather than repeating. What `amend` composes
is a different artefact on purpose: no task id, because a `task/NNNN-`
branch would make the machinery read the amendment as flight.

The report's coverage claim is wrong the other way round.
`tests/amend_lib.sh` and `tests/author_lib.sh` fixture the **shipped**
template, and `the_kits_own_check_accepts_the_body_test.sh` hands the
Go-composed body to `check_amendment_reference.sh` and requires it to
pass. The composers read the adopter's template at runtime; only the
sentinels are pinned, and pinning is the stated relationship.

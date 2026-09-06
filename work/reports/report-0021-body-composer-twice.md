---
id: report-0021
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T01:29:55Z
triaged: null
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

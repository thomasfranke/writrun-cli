---
id: task-0024
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0023]
doc_ref: null
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-06T04:28:21Z
queued: 2026-09-06T05:08:10Z
completed: 2026-09-06T08:38:24Z
merged: null
provenance: []
---

# Judge the composed title where the composition happens, not at the door

**References:** [spec-0023](../specs/spec-0023-observance-at-composition.md)

`writrun amend --type` accepts any lowercase word, so `--type wibble`
composes a title the project's own door refuses — and the refusal
arrives on the forge, after the branch is pushed and the pull request is
open. `product/pull-requests/shape.md` says a non-zero check stops the
command there, with no branch created and nothing left behind; `amend`
runs no repository check at all.

`author` has the same hole and a wider one: it runs three kit checks and
`check_observance.sh` is not among them, then asks the human for a
free-text title and pushes.

Validating locally is not the reimplementation the rules forbid. This
binary already reads the kit's own vocabulary in three places — the
commit-msg hook `init` installs extracts `TYPES=` from
`check_observance.sh` at commit time "so the hook and the door can never
disagree", `initcmd` writes that same field at adoption, and
`take_task.sh` reads it before composing. What is missing is the fourth:
handing the script the composed title and letting it answer.

`take` is unaffected — it delegates to `take_task.sh`, which runs
`valid_summary` before composing. `finish` composes no title.

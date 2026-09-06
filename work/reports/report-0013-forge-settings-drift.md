---
id: report-0013
status: declined
task_ref: []
doc_ref: null
created: 2026-09-05T18:21:48Z
triaged: 2026-09-06T00:38:00Z
---

# The forge settings this repository runs on do not match what stage 2 assumes

`writrun doctor`, run against this repository while it was being built,
reported two stage-2 findings. `gh api
repos/thomasfranke/writrun-cli/actions/permissions/workflow` answers
`{"default_workflow_permissions":"read"}`, where
`docs/product/adoption/doctor.md` states read-and-write. `gh api
repos/thomasfranke/writrun-cli/rulesets/22247734` — the `protect-main`
ruleset, `enforcement: active`, targeting `~DEFAULT_BRANCH` — answers
`"bypass_actors": []`, where the same document expects the Actions bot
on the bypass list. The recording pushes land all the same: each of the
four workflow files declares `permissions: contents: write` of its own,
and the ruleset's rules are `deletion`, `non_fast_forward`, `creation`
and `required_linear_history`, none of which refuses a fast-forward
push. Repository settings are a human gate in `AGENTS.md`, so nothing
was changed.

**Declined: the repository is right and the document is wrong.**
`doctor.md` asks for read-and-write "so the recording bot can push to
`main`", and the bot already pushes — `github-actions[bot]` carries 52
commits on `main`. Each workflow that writes raises its own
`contents: write`; the five that do not write stay on `read`. Setting
the repository default to read-and-write would hand write to all of
them, including `tests.yml`, to satisfy a document that fixed one means
to an end rather than the end. The bypass list is the same: the
`protect-main` rules are `deletion`, `non_fast_forward`, `creation` and
`required_linear_history`, none of which a fast-forward push meets, so
bypassing them grants a right nothing needs.

Declining destroys nothing: [[task-0020]] carries what this found — that
the stage-2 check reads a configuration where it should read a
capability.

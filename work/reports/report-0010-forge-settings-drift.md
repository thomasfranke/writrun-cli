---
id: report-0010
status: open
task_ref: []
doc_ref: null
created: 2026-09-05T18:21:48Z
triaged: null
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

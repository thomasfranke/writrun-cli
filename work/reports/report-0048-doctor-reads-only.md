---
id: report-0048
status: tracked
task_ref: [task-0039]
doc_ref: null
created: 2026-09-19T21:28:44Z
triaged: 2026-09-19T21:41:29Z
---

# doctor reads only rulesets, so a classic branch protection rule reports as no rule at all

**References:** [task-0039](../tasks/task-0039-classic-protection.md)

Issue #168, opened by @thomasfranke.

### What was observed

`doctor`'s stage-2 forge checks read only `/repos/{owner}/{repo}/rulesets`. They never read `/repos/{owner}/{repo}/branches/{branch}/protection`, so a branch governed by a classic branch protection rule is examined as if nothing governed it.

On a repository where `main` carries a classic rule and no ruleset targets it, stage 2 reports:

```
✓  the recording push can write to main
!  main is governed by a ruleset — no ruleset governs it — the methodology recommends protecting it; nothing blocks the recording push meanwhile
✓  no rule over main refuses the recording push
```

Three lines, one blind source. Two of them are stated as met and are not:

- the classic rule carries `required_pull_request_reviews` with `enforce_admins: true`, which refuses a direct push to `main` from every actor, `GITHUB_TOKEN` included;
- that is the one rule `writrun-approve.yml` names as the shape the recording cannot survive, and the repository is owner type `User`, where the forge offers no Actions bypass;
- so the recording push will be refused on the first merge, while doctor says twice that nothing refuses it, and grades the finding it did raise as recommended.

The finding's own line also reads as a contradiction against itself: `main is governed by a ruleset — no ruleset governs it`. The expectation and the observation are rendered into one sentence with no seam between them, so the line asserts and denies the same fact before reaching its advice.

### Evidence

Against `thomasfranke/tom` — public, `owner_type: User`, default branch `main`.

`gh api repos/thomasfranke/tom/rulesets` — one ruleset, and it does not target the branch:

```json
[{"id":20218584,"name":"Release tags","target":"tag","enforcement":"active"}]
```

`gh api repos/thomasfranke/tom/branches/main/protection` — the rule doctor never reads:

```json
{"required_pull_request_reviews":{"required_approving_review_count":0,
  "dismiss_stale_reviews":false,"require_code_owner_reviews":false},
 "enforce_admins":{"enabled":true},
 "required_linear_history":{"enabled":true},
 "allow_force_pushes":{"enabled":false},
 "allow_deletions":{"enabled":false},
 "required_conversation_resolution":{"enabled":true}}
```

`required_status_checks` is not enabled (404).

The shape named as unsurvivable, from the kit's own `.github/workflows/writrun-approve.yml`:

> a full ruleset with the pull-request requirement — the push lands only with GitHub Actions on the bypass list, which the forge offers on organization-owned repositories alone.

Stage 2 otherwise passes 5 of 6; the full run reports `1 finding at stage 3, none breaking a flow: 1 recommended`.

### Version consumed

`.writrun/VERSION` — v0.0.09, via `writrun-cli v0.0.3`.


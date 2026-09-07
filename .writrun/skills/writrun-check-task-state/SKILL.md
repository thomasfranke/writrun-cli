---
name: writrun-check-task-state
description: Use this skill before opening a pull request in a project that follows the WritRun methodology, or when verifying that a change's task and spec status transitions are legal — when the user asks to check task state, validate a lifecycle transition, or confirm a PR is not approving its own spec. Rejects the transitions the human gates exist to prevent.
---

# Check task state

Reads the task and spec status transitions a diff makes against the
lifecycle — *what states* a change moved files through, where
[`writrun-check-spec-deltas`](../writrun-check-spec-deltas/SKILL.md)
reads *which files* it touched. It is a script because the gate it holds
is `draft → approved`, the one transition an agent may never make,
including on a spec it drafted itself: asking that agent whether it
respected the gate is asking the wrong party.

## Run it

```bash
bash .writrun/skills/writrun-check-task-state/check_state.sh [diff-range]
```

The range defaults to `main...HEAD`. Two rules have two halves each, and
in both the second half needs an input the range does not carry.

The `tracked` route's: the **branch half** reads `HEAD_REF` when set (CI
passes the head branch that way) and the checkout's own branch
otherwise; on a detached HEAD with neither it says on stdout that it
stood down rather than passing quietly. The **diff half** reads what the
change carries, which is always there, so it runs on a detached HEAD too
— a change carrying code outside `work/` is refused whatever the branch
is called, and where it has no name at all.

The owed spec's: the **file half** reads the range alone, at every
stage, and refuses a task born `blocked` with a null `blocked_reason` —
a hold naming nothing to wait for is a hold no reader can release. The
**declaration half** reads `PR_BODY` (CI passes it the same way) and
refuses a newly added `origin: report` task that lands `backlog` with
`spec_ref: []`, no spec added beside it, and no line saying none is
warranted. Without a body it stands down on stdout and names the task it
could not judge. An *empty* body is not an unreadable one: it was read
and declares nothing, which is a refusal.

- **0 / OK** — no forbidden transition.
- **1** — every violation prints, each with the fix. The verdicts are
  `FORBIDDEN` (a gate), `INCONSISTENT` (a completion half-written) and
  `BROKEN` (a `spec_ref` resolving to no file); the lifecycle they are
  read against is
  [`stage-2-pull-requests/statuses.md`](https://github.com/thomasfranke/writrun/blob/main/docs/product/stage-2-pull-requests/statuses.md).
- **3** — usage error or `git diff` failed.

**Run it after the completion edits, never before.** Every rule it has is
about a transition, and the transitions it exists to reject are the ones
those edits make — run it earlier and it passes without reading anything.
[`preflight.sh`](../../scripts/stage-1-tasks-and-specs/preflight.sh) is
the ordered form of that rule and the way to run all three gates.

## Never

- Never satisfy `draft -> approved` by having a human approve the spec
  verbally and writing the field anyway. The gate is satisfied by a
  recorded approval of the change, and by nothing else.
- Never satisfy an `INCONSISTENT` result by blanking the `completed`
  date — finish the spec's Outcome instead.
- Never hand-edit the authority branch afterwards to undo a `FORBIDDEN`
  status move. The machinery writes that line from forge events; if it
  reads wrong, the event is what is missing.
- Never turn a `tracked` verdict into `fixed` to clear it: `fixed` says
  the change in hand ended the finding, and claiming that of one that
  still needs work loses it.
- Never try to clear a `tracked` verdict by **renaming the branch** to
  `report/…`. It no longer clears the check — rule K reads what the
  change carries as well as what it is called, and an implementing
  change carries code whatever its branch is named. What the rename
  still costs is the ride it was taken for: `apply_pr_event.sh` records
  every task the branch or the title names, so a rename that also drops
  the `[TASK-NNNN]` tag stops recording the task at all. Move the
  report, the task and the spec to a change of their own.
- Never skip it because the change touched no code. A change that only
  edits front matter is exactly what it is for.
- Never clear the owed-spec refusal by writing "No spec for task-NNNN"
  over a task whose spec is coming. The line says the task warrants
  none, permanently; a task waiting for one lands `blocked` with its
  reason, and the change that adds the spec releases it.

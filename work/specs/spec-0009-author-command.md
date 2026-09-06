---
id: spec-0009
task_ref: task-0010
status: implemented
created: 2026-09-03T22:30:41Z
---

# spec-0009 — Derived work presented, authoring PR opened ready

**References:** [task-0010](../tasks/task-0010-author-command.md)

- **Goal:** `writrun author` opens the authoring pull request for a finished rule.

## Scope

In: checks, branch and title composition, the Derived-work body, the
ready (not draft) pull request. Out: deriving the tasks and specs — the
agent's, upstream of this command; approving anything.

## Steps

1. Require a diff touching `docs/` plus derived queue files; run the checks in the fixed order (`check_front_matter.sh`, `check_doc_shapes.sh`, `check_state.sh`); stop at the first non-zero.
2. Compose: branch `docs/<short-name>`; title in the declared style, no task tags; body from the template's Derived-work half, the table filled from the tasks and specs the diff adds — or `none` declared when the rule derives nothing.
3. Show branch, title, body, files; on confirmation push and open the pull request ready.

## Acceptance criteria (EARS)

- When a check fails, the system shall stop there: no branch, no push, no pull request.
- When the diff adds tasks and specs, the body's Derived-work table shall list every one of them.
- When the rule derives nothing, the body shall declare none under `## Derived work`.
- When opened, the pull request shall be ready, not draft, and its title shall carry no task tag.

## Edge cases

- Diff already on a pushed branch: refuse; authoring starts locally.
- Mixed diff (docs plus unrelated code): refuse — one kind per change.

## Tests required

Integration over a fixture adoption with stubbed `gh`.

## Definition of Done

- [x] The opened PR passes the kit's own `check_derived_work.sh`.
- [x] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

`writrun author` is `internal/command/authorcmd/`, registered in
`commands()` (`cmd/writrun/main.go`). No package outside it changed.

- The diff is read against `origin/main...HEAD`, then `main...HEAD`;
  `--range` overrides. An empty diff, a diff touching no `docs/` path, a
  diff touching a path outside `docs/` and `work/`, a dirty tree, a
  detached HEAD and a branch already on the forge are refused before any
  check runs.
- The checks run as `check_front_matter.sh`, `check_doc_shapes.sh`,
  `check_state.sh <range>`. The first non-zero exit is returned unedited
  and nothing after it runs.
- The branch is `docs/<short-name>`: `--slug`, else the current branch
  when it already is a `docs/` one, else the doc the rule was written
  into. The title is the human's, asked in the style
  `stage_2.pr_title_style` names; a title carrying `[TASK-NNNN]` is
  refused and none is ever added.
- The body is `.writrun/templates/pull_request_template.md` with the
  `## Spec` half dropped and `## Derived work` filled from the tasks and
  specs the diff adds, or `none — this rule derives no work.` when it
  adds none. `gh pr create` runs without `--draft`.

Two things diverge from the steps above.

- The command carries `--resume`, which lifts the pushed-branch refusal.
  `product/rules.md` requires a failure after the first write to name the
  command that resumes the flow, and a branch pushed without its pull
  request is that state.
- Filling `## Derived work` drops the template's instruction comment from
  that section. The comment contains the word `none`, which
  `check_derived_work.sh` greps the section for, so a body keeping it
  would pass the check while declaring nothing
  (`work/reports/report-0016-derived-work-comment.md`).

Verification: `make tests` — every Go package `ok`, 111 bash case files
passed, 0 failed. `make cover` — 97.9% over `internal/` (floor 90%),
`internal/command/authorcmd` 98.3% (floor 80%). Six case files under
`tests/integration/author/` run the real kit scripts over a fixture
adoption with a bare local origin and a stubbed `gh`; two of them run
`check_derived_work.sh` over the body that was opened.

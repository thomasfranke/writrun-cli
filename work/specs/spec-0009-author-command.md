---
id: spec-0009
task_ref: task-0010
status: draft
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

- [ ] The opened PR passes the kit's own `check_derived_work.sh`.
- [ ] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

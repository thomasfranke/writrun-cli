---
id: spec-0011
task_ref: task-0012
status: approved
created: 2026-09-03T22:30:43Z
---

# spec-0011 — Back to draft with the suspended PR named

**References:** [task-0012](../tasks/task-0012-amend-command.md)

- **Goal:** `writrun amend` reopens an approved spec's gate, cross-referenced.

## Scope

In: returning a named approved spec to `draft`, composing the amendment
pull request, naming the suspended in-flight pull request. Out:
re-approving (the merge's, human-assented); touching any task status.

## Steps

1. Refuse a spec that is not `approved`.
2. Write `status: draft` on the spec — the one queue edit this command makes.
3. Find in-flight tasks (`in-progress`/`in-review`) referencing the spec and their open pull requests; compose the body line the reference check accepts: `Suspends #<n> — <task-id> waits on this amendment.`
4. Compose branch (id-less, `docs/` or type-prefixed), title in the declared style with no task tags; show everything; on confirmation push and open ready.

## Acceptance criteria (EARS)

- When the named spec is not approved, the system shall refuse.
- When an in-flight task references the spec, the opened body shall carry a reference `check_amendment_reference.sh` accepts.
- When amend completes, no task's `status` or `taken_by` shall have changed.
- When the forge cannot be read, the system shall say the reference must be checked by hand and still compose it from the queue.

## Edge cases

- Spec referenced by no in-flight task: the ordinary pre-implementation amendment; no reference owed.
- Several in-flight tasks on one spec: one Suspends line per pull request.

## Tests required

Integration with `WRITRUN_PR_LIST` fixtures covering suspended and
unsuspended cases.

## Definition of Done

- [ ] The opened PR passes the kit's own `check_amendment_reference.sh`.
- [ ] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

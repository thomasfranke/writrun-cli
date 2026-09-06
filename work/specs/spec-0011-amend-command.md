---
id: spec-0011
task_ref: task-0012
status: implemented
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

- [x] The opened PR passes the kit's own `check_amendment_reference.sh`.
- [x] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

`internal/command/amendcmd/` runs the four steps and decides nothing the
methodology had not: a spec that is not `approved` is refused before
anything else happens; the tasks in flight on it and the pull requests
working them are read the way `check_amendment_reference.sh` reads them
— `status` and `spec_ref` out of `work/tasks/*.md`, `ql_carried_of`'s
rule over each open pull request's head branch and title; the branch is
id-less (`docs/<slug>`, or `<type>/<slug>` under `--type`), the title is
composed in the style `read_setting.sh` reports and carries no task tag,
and the body is the adopter's own template with the authoring half and
the guidance comments dropped and the suspension lines put where the
gate reads them. `--yes` answers the one question; a yes writes
`status: draft`, commits `docs(specs): return spec-0011 to draft`,
pushes and opens the pull request **ready**. Coverage: 92.4% of the
package, 97.1% over `internal/`; 57 Go tests in the package and 43
assertions across eight integration cases under
`tests/integration/amend/`, one of which hands the composed body to the
kit's own `check_amendment_reference.sh` and asserts it passes — and
that the same gate over the same change rejects a body naming nothing.

**Where the one queue edit lands, and why that is still this order.**
The Steps put the write at step 2 and the confirmation at step 4, and
every effect here lands on the confirmed path instead. Nothing after
step 2 reads what step 2 writes: the suspended tasks, the open pull
requests, the branch, the title and the body are all read out of the
state as it stands *before* the amendment, so holding the edit and
writing it beside the push preserves the order of effects exactly while
letting a declined `amend` leave the working tree untouched — which is
what `product/pull-requests/shape.md` demands of a refused command ("no
half-written status, no orphan branch") and what report-0015 records
`finish` cannot do, because its step 4 reads its step 2. This is a
reading of the Steps, not a change to them, and it is stated here so a
reviewer can reject it consciously; `a_decline_leaves_nothing_behind`
asserts `git status --porcelain` is empty afterwards and that neither
the origin nor the forge saw anything. The judgement report-0015 asks
for is task-0018's, and this settles nothing on `finish`'s behalf.

Five things the Steps did not say, decided here and left visible: **the
branch is cut from a freshly fetched `origin/main`** (falling back to the
local `main`, then `HEAD`), and the spec is **re-read on that branch**
before it is written, so an amendment never pastes the old checkout's
bytes over a newer spec — and a base whose spec is no longer `approved`
stops the run rather than overwriting somebody else's amendment; **a
dirty tree is refused** up front, because it would ride into the branch;
**the title's sentence is the human's** (`--title`, or typed) while the
labels and the scope are composed, the scope being `specs` since
`work/specs/` is the whole of what an amendment touches; **the commit
subject stays Conventional Commits** whatever the title style says, per
`conventions/commits.md`; and **a forge that will not answer is
best-effort**, the contract `check_amendment_reference.sh` itself states
— the command says the reference must be checked by hand, composes it
from the queue naming the task instead of a number, and opens anyway.

`WRITRUN_PR_LIST` is honoured as the kit's own seam, so the suite
answers the forge's question without a forge; `gh` is stubbed and the
origin is a bare local repository in every case. Every helper stayed
inside the package — `internal/kit`, `internal/forge`, `internal/gitx`,
the frame's `Ctx`/`Terminal`, `finishcmd` and `takecmd` are untouched —
so the front-matter reader, the queue-file resolver and the id parser
are duplicated with `finishcmd` on purpose, for task-0019 to
consolidate. No task's `status` or `taken_by` is written anywhere in the
package, and no `docs/` path changed: this spec promised none.

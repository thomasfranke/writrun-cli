---
id: spec-0005
task_ref: task-0005
status: approved
created: 2026-09-03T22:30:37Z
---

# spec-0005 — Remove what init installed, keep the project's record

**References:** [task-0005](../tasks/task-0005-uninstall-command.md)

- **Goal:** `writrun uninstall` removes the kit; the project's record survives.

## Scope

In: removing kit files, the commit-message hook init installed, and the
fenced `AGENTS.md` section. Out — never touched: `work/`, the project's
docs, anything outside the fence.

## Steps

1. Refuse where `.writrun/` is absent.
2. Compute the removal set (`.writrun/`, `WRITRUN.md`, `docs/writrun-instructions.md`, the four workflows, the hook, the fenced section) and the keep set (`work/`, docs, the rest of `AGENTS.md`).
3. Show both sets; confirm; remove.

## Acceptance criteria (EARS)

- When run where `.writrun/` is absent, the system shall refuse.
- When uninstall completes, `work/` and the project's docs shall be byte-identical to before.
- When `AGENTS.md` holds content outside the fence, that content shall survive byte-identical; only the fenced section is removed.
- When the installed hook is not the one init writes, the system shall leave it and say so.

## Edge cases

- `AGENTS.md` that is pure kit skeleton: removed whole, named in the shown set.
- Kit files already partially deleted by hand: remove what remains, list what was already gone.

## Tests required

Integration: adopt, then uninstall; assert the keep set untouched and
the removal set gone.

## Definition of Done

- [ ] Uninstall after a real adoption leaves only the project's files.
- [ ] Suite green.

## Proposed product changes

- none — `product/adoption/uninstall.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

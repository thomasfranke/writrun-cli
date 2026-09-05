---
id: spec-0005
task_ref: task-0005
status: implemented
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

`writrun uninstall` ships in `internal/command/uninstallcmd/`. The
removal set is `internal/kitpaths`, the same inventory update refreshes
from; the fenced section is cut by `internal/fence.Remove`, which gives
back the blank line the graft added, so a graft followed by a removal
is a byte-for-byte round trip.

The hook moved to `internal/hook` so that removing it is a recognition
rather than a guess: `Inspect` reports `Ours` only on a byte-identical
match with the script init writes, and anything else is `Foreign` and
left standing with its own line in the shown plan. A hook a project
edited is a hook a project owns.

Verified by `tests/integration/uninstall/` — five cases: the full
removal with the queue and a project chapter asserted byte-identical
and AGENTS.md restored to its pre-adoption bytes, the foreign hook left
behind, content surviving on both sides of the fence, the bare skeleton
removed whole, and the refusal where `.writrun/` was never installed.
`make tests` exits 0.

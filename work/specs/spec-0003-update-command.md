---
id: spec-0003
task_ref: task-0003
status: implemented
created: 2026-09-03T22:30:36Z
---

# spec-0003 — Refresh only what the methodology declares refreshable

**References:** [task-0003](../tasks/task-0003-update-command.md)

- **Goal:** `writrun update` refreshes the kit without touching what is the project's.

## Scope

In: refreshing `skills/`, `scripts/`, `templates/`, `VERSION`, the four
workflows, and the fenced `AGENTS.md` section. Out — never touched:
`conventions/`, `settings.json`, the project's docs, `work/`.

## Steps

1. Require an adopted repository; read the current tag from `.writrun/VERSION`; fetch the target tag as init does.
2. Diff kit-owned paths; show what will change before changing it.
3. Refresh: replace kit-owned directories whole; rewrite `VERSION`.
4. `AGENTS.md`: replace only the content between the markers, preserving the two lines marked `yours` (the gates table, the deriving default).
5. Stop, changing nothing, if either marker is missing or damaged.

## Acceptance criteria (EARS)

- When the fenced markers are missing or damaged, the system shall stop and change nothing.
- When update completes, `conventions/`, `settings.json`, `docs/` and `work/` shall be byte-identical to before.
- When update completes, the lines marked `yours` shall survive inside the refreshed section.
- When the target tag equals the recorded tag, the system shall say so and change nothing.

## Edge cases

- Hand edits inside kit-owned folders: overwritten by design — shown in the diff first.
- Target tag older than recorded: refuse; a downgrade is a deliberate act the command does not offer.

## Tests required

Integration: adopt at one tag, update to the next, assert the
untouchable set byte-identical and the `yours` lines preserved.

## Definition of Done

- [ ] Update between two real WritRun tags passes all criteria.
- [ ] Suite green.

## Proposed product changes

- none — `product/adoption/update.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

`writrun update` ships in `internal/command/updatecmd/`. The refresh set
is read from `internal/kitpaths`, the inventory both this command and
uninstall share; the fenced section is rewritten by `internal/fence`,
whose `Replace` carries every block a `yours` marker governs across —
the marker sits before the gates table and after the deriving default,
so a block is taken from whichever side of the marker holds one.

Two behaviours the spec did not name, both refusals:

- **A section carrying fewer `yours` markers than the document is
  refused.** The steps say the marked lines survive; a tag that dropped
  a marker would take the project's answer with it silently, which is
  the one outcome the step forbids.
- **A dirty tree is refused**, as init refuses one. A refresh
  *overwrites*, so an uncommitted edit inside a kit-owned folder would
  be gone with nothing to restore it from.

Verified by `tests/integration/update/` — four cases over two real tags
in the fixture source: the refresh with the untouchable set asserted
byte-identical and both `yours` blocks preserved, the same-tag
stand-down, the damaged fence, and the refused downgrade.
`writrun-check-spec-deltas spec-0003,spec-0005` exits 0; `make tests`
exits 0.

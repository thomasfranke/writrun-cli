---
id: spec-0033
task_ref: task-0030
status: draft
created: 2026-09-08T16:15:40Z
---

# spec-0033 — The adopter's files move once, and the pin follows

**References:** [task-0030](../tasks/task-0030-two-homes-catchup.md)

- **Goal:** `writrun` names the adopter's files at `writrun/`, moves a
  repository still on the old layout there once, and pins WritRun
  `v0.0.06`.

## Scope

In: the two addresses the binary reads, the paths a refresh must not
overwrite, the one-time move, and the pinned tag.

Out: deciding the migration's behaviour. WritRun's spec-0090 wrote that
contract and put this client outside its own scope; what is implemented
here is what it states, and a disagreement is resolved in its favour
([about](../../docs/about.md)).

Also out: `uninstall`'s reach. Under the two homes a removal deletes
`.writrun/` and the named files and leaves `writrun/` standing, which
is what `uninstall` already promises for the adopter's paths — but the
paths it recognises are this spec's to repoint, not to redesign.

## Steps

1. Repoint `kit.Settings` and `kit.Gates` in `internal/kit/kit.go` to
   `writrun/settings.json` and `writrun/gates.md`. Task-0027 put every
   kit path in that one file, so this is the whole of the renaming and
   no second copy can drift behind it.
2. Repoint `kitpaths.Untouchable` and `kitpaths.Seeded` to the new
   addresses, and add `writrun/conventions`. **This is the load-bearing
   edit:** while `Untouchable` names the old addresses, a refresh treats
   the adopter's new files as the kit's and writes over them.
3. Implement the move, in `update`, before the refresh writes anything:
   where `.writrun/settings.json`, `.writrun/gates.md` or
   `.writrun/conventions/` exist and the matching `writrun/` path does
   not, move them, preserving content byte for byte.
4. Where both addresses hold a file, keep the new one, leave the old one
   on disk, and name it in the report. Nothing is merged.
5. Show the move in `update`'s plan, before its confirmation — a
   migration is a change to the repository, so it is shown and asked
   for like every other ([rules](../../docs/product/rules.md)).
6. Move the pin to `v0.0.06`.
7. Repoint the prose that names the old addresses: the findings in
   `doctorcmd` and `initcmd`, `amendcmd`'s comment, and `init`'s plan
   line.

## Acceptance criteria (EARS)

- When the binary reads a setting or a gate, the system shall read it
  from `writrun/`.
- When a refresh runs, the system shall not write any path under
  `writrun/`.
- When `update` runs where an adopter file sits at its old address and
  the new address is free, the system shall move it and shall preserve
  its content byte for byte.
- When `update` runs where both addresses hold the file, the system
  shall keep the file at the new address, shall leave the old one
  untouched, and shall name both in its report.
- When `update` would move a file, the system shall show the move in
  its plan before asking for confirmation.
- When the confirmation is declined, the system shall move nothing.
- When `update` runs where the layout is already the new one, the system
  shall move nothing and shall report nothing about a migration.
- When `--version` runs, the system shall name `v0.0.06` as the tag it
  pins.
- When `init` adopts a repository, the system shall write the adopter's
  files at the new addresses only.

## Edge cases

- **A repository at the old layout whose `writrun/` exists for another
  reason.** The rule is per file, not per folder: a `writrun/` holding
  something unrelated does not block moving a file whose own new
  address is free.
- **`.writrun/conventions/` with files the kit never shipped.** The move
  carries the folder whole. Nothing selects within it, because what is
  inside is the adopter's by the two-homes rule.
- **The legacy `.writrun/conventions/settings.json`.** The kit's own
  `check_settings.sh` names that address and refuses to read it; this
  command does not move it either. One migration is the contract, and
  it is the one spec-0090 wrote.
- **A move that fails halfway.** The refresh has not begun, so the
  failure names what moved and what did not, and nothing of the kit is
  written — a half-migrated repository must not also be half-refreshed.
- **`doctor` run between the move and the refresh.** It reads the new
  addresses, which is where the files now are.

## Tests required

- Unit, `internal/kitpaths`: the three new paths are untouchable; the
  three old ones are not; a fixture asserting `writrun/settings.json`
  survives a refresh that ships one.
- Unit, `internal/command/updatecmd`: the move happens where the new
  address is free; both-present keeps the new file and reports the old;
  an already-migrated layout moves nothing; a declined confirmation
  moves nothing.
- Unit: the moved content is byte-equal to what was there.
- Integration, `tests/integration/update/`: a fixture repository seeded
  at the old layout, updated, ends with its own answers at the new
  addresses — the existing "refreshes the kit and keeps the project"
  case extended to the migration.
- The `adoption stage` and `settings are canonical` workflows must stay
  green against this repository's own migrated files.

## Definition of Done

- [ ] No live Go file names an adopter file under `.writrun/`.
- [ ] A refresh cannot write under `writrun/`.
- [ ] A repository on the old layout is migrated once, on the word.
- [ ] The pin reads `v0.0.06`.
- [ ] This repository's own `settings.json` and `gates.md` are at the
      new addresses, carrying the same values they carry today.

## Proposed product changes

- `product/adoption/update.md` — the adopter's paths are named at their
  new addresses, and the one-time move is stated with what it does
  where both exist.
- `product/adoption/init.md` — the settings file's address.
- `product/adoption/uninstall.md` — what a removal leaves standing,
  named as the folder rather than the file list.
- `product/rules.md` — the sentence naming the adopter's own files
  points at `writrun/`.

## Proposed technical changes

- `technical/engineering/coupling.md` — the two homes as the boundary
  the binary reads, and the rule that a path the kit owns is never
  written by the adopter's side.
- `technical/layout/tree.md` — `writrun/` in the tree.
- `technical/decisions/architecture/` — the decision recording that the
  client follows the kit's address contract rather than holding one of
  its own.

## Outcome

_(fill after execution)_

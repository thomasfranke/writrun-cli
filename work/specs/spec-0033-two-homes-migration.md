---
id: spec-0033
task_ref: task-0030
status: draft
created: 2026-09-08T16:15:40Z
---

# spec-0033 — The adopter's files move once, and the pin follows

**References:** [task-0030](../tasks/task-0030-two-homes-catchup.md)

- **Goal:** `writrun` reads the adopter's answers at `writrun/`,
  resolves a deferring file to the kit's default, moves a repository
  still on the old layout once, and pins WritRun `v0.0.07` — leaving a
  shape where `v0.0.08` is a constant to change, not a migration to
  design.

## Scope

In: the addresses the binary reads, how it reads a file that defers,
where the commit vocabulary is written, the paths a refresh must not
overwrite, the one-time move, and the pinned tag.

Out: deciding any of it. WritRun's `spec-0090` and `v0.0.07`'s two-homes
rule wrote the contract and put this client outside their own scope;
what is implemented here is what they state, and a disagreement resolves
in their favour ([about](../../docs/about.md)).

## The shape being followed

`v0.0.07` layers the two homes. Every text the kit answers on a
project's behalf ships as a default under `.writrun/defaults/`, and the
project's home carries one file per answer that either **defers** or
**overrides whole**. Deferring is positional: a first line of
`/// writrun:default`, and there only. A file carrying it — or absent —
is the kit's answer; anything else is the project's, whole. Nothing
merges, and `writrun/` is still a home an update never touches: the kit
corrects a default by replacing its own home, and the correction reaches
every project that defers.

Two consequences this binary must follow. A reader resolves through
`.writrun/scripts/stage-1-tasks-and-specs/resolve_doc.sh` rather than
reading the project's file directly. And the commit vocabulary is
`stage_2.commit_types` and `stage_2.commit_scopes` in
`writrun/settings.json` — the checks read it there alone, so no prose
carries a second copy for a check to disagree with.

## Steps

1. Repoint `kit.Settings` and `kit.Gates` to `writrun/settings.json` and
   `writrun/gates.md`, and add `kit.ResolveDoc` for the resolver.
2. Repoint the four call sites that build an adopter path from segments
   rather than reading a constant, all in `initcmd`: `plan.go`'s
   settings writer, `checks.go`'s two readers, and `conventions.go`'s
   `commits.md` reader. **They are live duplicates** — task-0027
   consolidated the paths the binary runs as scripts, and
   `filepath.Join(root, ".writrun", …)` escaped its reach. Repointing a
   constant alone would compile and leave `init` writing to the old
   address.
3. Leave the `.writrun` references that are not the adopter's: `wrepo`'s
   adoption probe, `kitpaths.RemoveDirs`, `updatecmd`'s refresh roots,
   and `conventions.go`'s `check_observance.sh`. The kit's home does not
   move.
4. Read every deferring file through the resolver, never by opening the
   project's path. `doctor`'s gates check is the case that breaks
   otherwise: a `writrun/gates.md` that defers holds no table, and the
   check would report every gate unanswered when the default answers
   them all.
5. Write the extracted commit vocabulary to `stage_2.commit_types` and
   `stage_2.commit_scopes`, and stop writing it into
   `conventions/commits.md` and `check_observance.sh`. The second was a
   kit file a refresh replaces, so an extracted vocabulary did not
   survive an update.
6. Point `kitpaths.Untouchable` at `writrun` — the home entire, because
   no folder is part of both — and **empty `kitpaths.Seeded`**. A
   refresh writes nothing into the project's home. **This is the
   load-bearing edit:** while `Untouchable` names the old addresses, a
   refresh treats the adopter's files as the kit's and writes over them.
7. Implement the move, in `update`, before the refresh writes anything:
   where an adopter file sits at its old address and the matching
   `writrun/` path is free, move it, preserving content byte for byte.
8. Where both addresses hold a file, keep the new one, leave the old one
   on disk, and name both in the report. Nothing merges.
9. Show the move in `update`'s plan, before its confirmation — a
   migration changes the repository, so it is shown and asked for like
   every other ([rules](../../docs/product/rules.md)).
10. Move the pin to `v0.0.07`.
11. Repoint the prose that names the old addresses: the findings in
    `doctorcmd` and `initcmd`, `amendcmd`'s comment, and `init`'s plan
    line.
12. Guard the shape, so the next tag is a constant and not a migration:
    a test that fails when a live Go file names an adopter path outside
    `internal/kit`, by literal or by segment. Task-0027 consolidated the
    script paths and four adopter paths escaped it precisely because
    nothing failed when they did.

## This repository's own target state

Decided by the maintainer, 2026-09-08: **every text the kit answers on
this project's behalf defers.** `writrun/gates.md` and every file under
`writrun/conventions/` carry the stub; only `writrun/settings.json`
keeps this project's values, plus `commit_types` and `commit_scopes`
carrying the vocabulary its `conventions/commits.md` states today.

The audit that supports it: of the seven conventions files here, six are
not customizations but stale kit text — `branches.md` is identical,
`README.md` and `specs.md` differ only in links the kit has since
corrected, and `prs.md`, `tasks.md` and `commits.md` are earlier
versions of the kit's own paragraphs. Only `prose.md` was rewritten on
purpose, and `gates.md` carries one row the kit's default lacks. Both
are given up deliberately: the docs law `prose.md` points at is stated
in `AGENTS.md` and `.ai/skills/docs/SKILL.md`, which no update touches,
and the routing gate is stated in the kit's own `.writrun/AGENTS.md` —
"ask the user, per report and never assumed from the conduct flags" —
so the row was a second copy of a rule every project already receives.

## Acceptance criteria (EARS)

- When the binary reads a setting or a gate, the system shall read it
  from `writrun/`.
- When the binary reads a file that defers, the system shall resolve it
  through the kit's resolver and shall judge the default's content.
- When `doctor` runs against a repository whose `writrun/gates.md`
  defers, the system shall report the gates as answered.
- When `init` extracts a commit vocabulary, the system shall write it to
  `stage_2.commit_types` and `stage_2.commit_scopes` and to no other
  file.
- When a refresh runs, the system shall not write any path under
  `writrun/`.
- When a refresh runs and a file under `writrun/` is absent, the system
  shall not create it — seeding the project's home is adoption's act.
- When a refresh runs, the system shall replace every file under
  `.writrun/` entire, without preserving a hand edit.
- When `update` runs where an adopter file sits at its old address and
  the new address is free, the system shall move it and shall preserve
  its content byte for byte.
- When `update` runs where both addresses hold the file, the system
  shall keep the file at the new address, shall leave the old one
  untouched, and shall name both in its report.
- When `update` would move a file, the system shall show the move in its
  plan before asking for confirmation.
- When the confirmation is declined, the system shall move nothing.
- When `update` runs where the layout is already the new one, the system
  shall move nothing and shall report nothing about a migration.
- When `--version` runs, the system shall name `v0.0.07`.

## Edge cases

- **A first line of `/// writrun:default` in a file that also carries
  prose.** The marker decides, and it decides alone: the rule is
  positional, so the rest is the stub's own explanation and never a
  partial answer.
- **A repository at the old layout whose `writrun/` exists for another
  reason.** The rule is per file: a `writrun/` holding something
  unrelated does not block moving a file whose own new address is free.
- **A move that fails halfway.** The refresh has not begun, so the
  failure names what moved and what did not, and nothing of the kit is
  written — a half-migrated repository must not also be half-refreshed.
- **The resolver absent.** A `v0.0.04` kit ships none. The migration and
  the refresh are one command, so the resolver is present by the time
  anything reads through it; a read that finds none names the kit's
  version rather than guessing at a path.
- **The legacy `.writrun/conventions/settings.json`.** The kit's own
  `check_settings.sh` names that address and refuses to read it; this
  command does not move it either.
- **A hand-edited kit file.** It is overwritten, by design. Nothing
  detects or preserves it: a kit change worth having is a report routed
  upstream, not an edit held locally.

## Tests required

- Unit, `internal/kitpaths`: `writrun/` is untouchable whole; the old
  addresses are not; nothing is seeded.
- A guard over the source: no live Go file outside `internal/kit` names
  an adopter path, by literal or by `filepath.Join` segment. It is what
  keeps the next tag from being a migration.
- Unit, `internal/command/doctorcmd`: a deferring `gates.md` reports its
  gates answered; an overriding one is judged on its own table.
- Unit, `internal/command/initcmd`: an extracted vocabulary lands in the
  two settings keys and nowhere else.
- Unit, `internal/command/updatecmd`: the move happens where the new
  address is free; both-present keeps the new file and reports the old;
  an already-migrated layout moves nothing; a declined confirmation
  moves nothing; the moved content is byte-equal.
- Integration, `tests/integration/update/`: a fixture repository seeded
  at the old layout, updated, ends with its own answers at the new
  addresses and its conventions deferring.
- The `adoption stage` and `settings are canonical` workflows stay green
  against this repository's own migrated files.

## Definition of Done

- [ ] No live Go file names an adopter file under `.writrun/`.
- [ ] A refresh cannot write under `writrun/`.
- [ ] A deferring file is judged by the default it resolves to.
- [ ] The commit vocabulary lives only in the settings.
- [ ] A repository on the old layout is migrated once, on the word.
- [ ] The pin reads `v0.0.07`.
- [ ] A test fails when an adopter path is named outside `internal/kit`,
      so the next tag is a constant to change.
- [ ] This repository's `gates.md` and conventions defer; its
      `settings.json` carries the values it carries today.

## Proposed product changes

- `product/adoption/update.md` — the adopter's paths at their new
  addresses, and the one-time move with what it does where both exist.
- `product/adoption/init.md` — the settings address, and the vocabulary
  written as settings rather than as prose.
- `product/adoption/doctor.md` — a gate answered by a default is
  answered.
- `product/adoption/uninstall.md` — what a removal leaves standing,
  named as the folder rather than the file list.
- `product/rules.md` — the sentence naming the adopter's own files
  points at `writrun/`.

## Proposed technical changes

- `technical/layout/tree.md` — `writrun/` in the tree.

## Outcome

_(fill after execution)_

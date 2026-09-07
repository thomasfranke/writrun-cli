---
id: spec-0027
task_ref: task-0028
status: draft
created: 2026-09-07T03:15:46Z
---

# spec-0027 — The kit takes nothing and leaves nothing

**References:** [task-0028](../tasks/task-0028-adopter-kit-boundary.md)

- **Goal:** uninstall keeps every adopter-owned byte and removes every kit-owned one, init and doctor require of `docs/` only what the kit declares, and the boundary is stated once in `AGENTS.md`.

## Scope

In: `internal/command/uninstallcmd`, `internal/command/initcmd`,
`internal/command/doctorcmd`, `internal/kitpaths`, `internal/chapter`
(its removal), root `AGENTS.md`, and the five docs under Proposed
changes.

Out: the kit template's own shape — where the kit puts its files is
upstream's, and a change there is a report ending `routed`. Out:
`update`'s refresh semantics, correct since spec-0025. Out: the
machine-local `.claude/` surveys — untracked, deleted by hand.

## What is wrong

| Defect | Evidence |
|---|---|
| Uninstall deletes the adopter's answers | `kitpaths.RemoveDirs = [".writrun"]` removes `settings.json`, `gates.md` and `conventions/` — the paths `kitpaths.Untouchable` declares the adopter's. |
| Uninstall leaves kit residue | `kitpaths.RemoveFiles()` names `WRITRUN.md` and `docs/writrun-instructions.md` only; the kit-shipped `CLAUDE.md` shim stays. |
| The binary requires what the kit calls optional | `initcmd/checks.go:82–88` and `doctorcmd/files.go:41–48` require `docs/about.md`, a `docs/product/` chapter and a `docs/technical/` doc; the kit's `docs/writrun-instructions.md` states "WritRun prescribes nothing inside `docs/`", names `product/` and `technical/` "optional, a suggestion only", and requires the About "under any name". |
| No single boundary statement | Root `AGENTS.md` carries the routing rule and the kit skill, but no section states what the adopter edits and what it never edits. |
| Layout drift | `technical/layout/tree.md` still describes `internal/kitpaths/` as "What `init` installs, listed once" — the inventory spec-0025 removed. |

## Steps

1. **Uninstall preserves the adopter's answers.** Replace the
   `RemoveDirs` whole-tree delete with a walk of `.writrun/` that
   removes every entry except those `kitpaths.Untouchable` names
   beneath it. When they survive, `.writrun/` stays holding only them,
   and the closing message names them. Add `--purge` to remove
   `.writrun/` whole, behind the same confirmation.
2. **Adoption is the recorded tag, not the folder.** Key `init`'s
   refusal on `.writrun/VERSION` present rather than `.writrun/`
   present, so uninstall then `init` round-trips the adopter's answers
   byte-identical — `init`'s walk already keeps every existing path.
3. **No unedited kit entry file stays behind.** Uninstall best-effort
   fetches the tag `.writrun/VERSION` records via `internal/kitfetch`.
   A file the template ships outside `.writrun/` that is not already in
   the removal set and is byte-identical in the repository is removed;
   one that differs stays and is named among what stays; a failed fetch
   keeps them all, names each as unjudged, and does not fail the
   uninstall. `AGENTS.md` stays `internal/pointer`'s alone — the
   comparison never applies to it, and no shipped content enters Go
   ([coupling](../../docs/technical/engineering/coupling.md) rule 3).
4. **Enforce only the docs shape the kit declares.** Drop the
   `docs/about.md` fixed-path check and both chapter checks from
   `initcmd/checks.go` and `doctorcmd/files.go`. Replace them with one
   finding, raised only when `docs/` holds no file besides
   `kitpaths.KitOwned`, worded on the kit's About rule. Keep the
   `docs/`-and-`work/` split check — the kit states it. Delete
   `internal/chapter/` if nothing else uses it.
5. **State the boundary in `AGENTS.md`.** Rework "This repository is
   the CLI, and only the CLI" to carry the contract, linking rather
   than restating: kit files are never edited by hand — `writrun
   update` replaces them, and a kit change worth having is an upstream
   issue per the existing routing rule; customization lives only in the
   adopter area — `.writrun/settings.json`, `.writrun/gates.md`,
   `.writrun/conventions/`, `AGENTS.md`, `CLAUDE.md`, `docs/`, `work/`;
   this repository's rules cite its own `docs/`, never kit internals;
   the binary enforces only what the kit declares.
6. **Docs follow behaviour** — the five files under Proposed changes,
   including the enforcement rule added to
   [coupling](../../docs/technical/engineering/coupling.md): a
   requirement no kit file states is not the binary's to enforce.

## Acceptance criteria (EARS)

- When `writrun uninstall` runs and is confirmed, the system shall leave `.writrun/settings.json`, `.writrun/gates.md` and every file under `.writrun/conventions/` byte-identical in place, and remove every other entry under `.writrun/`.
- When `writrun uninstall --purge` runs and is confirmed, the system shall remove `.writrun/` whole.
- When uninstall runs and `CLAUDE.md` is byte-identical to the fetched template's copy, the system shall remove it; when it differs, the system shall keep it and name it among what stays.
- When uninstall cannot fetch the recorded tag, the system shall keep every entry-point file, name each as unjudged, and complete the removal with exit 0.
- When `writrun init` runs where `.writrun/VERSION` records a tag, the system shall refuse; where `.writrun/` holds only adopter-owned files, it shall install and leave those files byte-identical.
- When `writrun doctor` runs at stage 1 against a `docs/` holding at least one file the kit did not ship, the system shall report no docs-shape finding.
- When doctor runs against a `docs/` holding only `writrun-instructions.md`, the system shall raise one finding naming the kit's About rule.
- When production source is searched for `docs/about.md`, `docs/product` or `docs/technical` as required paths, the search shall find none.

## Edge cases

- `conventions/` already deleted by the adopter: nothing to preserve, and the uninstall proceeds.
- A stray adopter file inside `.writrun/` outside the `Untouchable` paths: the folder is documented kit territory, the file is removed, and the plan shows it before the confirmation.
- `AGENTS.md`: pointer-cut only, never template-compared — an adopter who rewrote everything around the pointer loses no byte outside the section.
- A `CLAUDE.md` that predates adoption: it differs from the template's copy, so it stays.
- A `gates.md` the adopter extended with rows of its own: `Untouchable`, so preserved verbatim; seeding never applies to a present file.

## Tests required

- `uninstallcmd`: preservation of the three adopter paths, `--purge`, shim byte-identical and shim edited, and the failed fetch, against a fake `kitfetch`.
- `initcmd`: refusal on a recorded tag; install over a preserved trio asserting byte-identity.
- `initcmd` and `doctorcmd`: the docs-shape checks rewritten to the kit's declared minimum.
- `tests/e2e/adopt`: the round-trip uninstall → `init` added to the parity suite.
- Every existing suite green.

## Definition of Done

- [ ] Uninstall keeps the three adopter paths and removes everything else of the kit's; `--purge` removes `.writrun/` whole.
- [ ] Uninstall then `init` round-trips the adopter's answers byte-identical.
- [ ] An unedited kit-shipped `CLAUDE.md` does not survive uninstall; an edited one does.
- [ ] No production check requires `docs/about.md`, a `docs/product/` chapter, or a `docs/technical/` doc.
- [ ] `AGENTS.md` states the boundary contract in one section, referencing the existing routing rule rather than restating it.
- [ ] The five docs under Proposed changes match behaviour; every suite passes.

## Proposed product changes

- `product/adoption/uninstall.md` — the adopter's answers survive; `--purge`; how entry files are judged against the fetched template.
- `product/adoption/init.md` — refusal keyed to `.writrun/VERSION`; re-adoption keeps preserved answers.
- `product/adoption/doctor.md` — the Stage 1 files bullet rewritten to the kit's declared minimum.

## Proposed technical changes

- `technical/engineering/coupling.md` — the enforcement rule: a requirement no kit file states is not the binary's to enforce.
- `technical/layout/tree.md` — `internal/kitpaths/` described by its post-spec-0026 role; the `internal/chapter/` row removed.

## Outcome

_(fill after execution)_

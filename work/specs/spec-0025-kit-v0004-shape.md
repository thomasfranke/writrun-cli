---
id: spec-0025
task_ref: task-0026
status: approved
created: 2026-09-06T20:35:04Z
---

# spec-0025 — The kit shape WritRun v0.0.04 ships

**References:** [task-0026](../tasks/task-0026-kit-v0004-shape.md)

- **Goal:** the adoption commands satisfy [coupling](../../docs/technical/engineering/coupling.md) against WritRun `v0.0.04`, and this repository's kit records that tag.

## Scope

In: `internal/kitpaths`; `internal/fence`, renamed `internal/pointer`;
`initcmd/plan.go` and `initcmd/checks.go`; `updatecmd/plan.go` and
`updatecmd/updatecmd.go`; `uninstallcmd/uninstallcmd.go`;
`doctorcmd/files.go`; the pin in `cmd/writrun/main.go`; this
repository's `.writrun/`, `.github/`, `AGENTS.md` and
`.ai/skills/docs/SKILL.md`.

Out: what the kit's scripts do, which is upstream's. Out: rewriting an
adopter's `AGENTS.md`, which `v0.0.04` makes the project's whole. Out:
`writrun init`'s conventions extraction, its stage question and its
hook, none of which reads the kit's inventory.

**The load-bearing premise**, from `template/.writrun/AGENTS.md` and
`template/.writrun/gates.md` at `v0.0.04`: the kit owns
`.writrun/AGENTS.md` and replaces it whole, the project owns
`.writrun/gates.md` and no update touches it, and everything above the
`## WritRun` heading of the root `AGENTS.md` is the project's. If that
is wrong, this spec's shape is wrong.

## What is wrong

The binary knows the kit by heart instead of reading it, in three
places, and `v0.0.04` makes each one wrong.

**The inventory.** `internal/kitpaths` names three refresh directories
and four workflows. Everything else the kit ships is refreshed by no
tag: `.writrun/README.md`, `WRITRUN.md` and `docs/writrun-instructions.md`
have stayed at whatever tag installed them since adoption. `v0.0.04`
adds `.github/workflows/writrun-intake.yml`,
`.github/ISSUE_TEMPLATE/writrun-report.yml`, `.writrun/AGENTS.md`,
`.writrun/gates.md` and `CLAUDE.md`, and a closed list means five files
arrive only where a Go change names them one by one.

`updatecmd/plan.go`'s apply loop compounds it: it writes a planned
change only where the path starts with `.github/`, on the reasoning that
the refresh directories carried every other one. A single kit file
outside both is counted in the plan and left on disk unchanged.

**The gates.** `doctorcmd/files.go` carries `theGates` — four
transitions and the phrase each is recognised by. `v0.0.04` states
seven, in `.writrun/gates.md`, a file the binary does not read. A kit
that adds an eighth needs a Go change to be checked.

**The fence.** `v0.0.04` carries no `writrun:begin` marker in
`template/AGENTS.md`, and six sites read one: `initcmd/plan.go:108`,
`initcmd/checks.go:99`, `updatecmd/plan.go:103`,
`updatecmd/updatecmd.go:89`, `uninstallcmd/uninstallcmd.go:126` and
`doctorcmd/files.go:75`. The first refusal is `update`'s, which stops
before the fetch and refreshes nothing.

Of the three, only the fence is a shape this binary should know: the
graft into an existing `AGENTS.md` is the one edit no kit file can
describe. The other two are the kit describing itself, and the binary
reciting it from memory.

## Steps

1. Reduce `internal/kitpaths` to what the kit does not declare: `Untouchable`, the adopter-owned paths a refresh never rewrites; `KitOwned`, the kit's own files that sit under one of them; and `uninstall`'s lists. `RefreshDirs`, `RefreshFiles` and `Workflows` go — a refresh reads the fetched template instead of a list.
2. Rewrite the refresh as one walk of that template: every file it ships is written unless an `Untouchable` path covers it and `KitOwned` does not carve it out; a file under a refreshed directory that the tag no longer ships is removed; every planned change is written, whatever its path.
3. Give the adopter-owned files inside `.writrun/` a seeding rule: absent, the template's copy is written once; present, it is never touched. Adopter-owned files at the repository root stay `init`'s alone — a refresh does not create the project's `AGENTS.md` or `CLAUDE.md`.
4. Have `uninstall` remove the kit's files by the namespace they carry — `writrun-*` under `.github/workflows/` and `.github/ISSUE_TEMPLATE/` — rather than by name, since it has no template to read.
5. Rename `internal/fence` to `internal/pointer` and give it the shape it now edits: `Section` cuts the heading whose body links `.writrun/AGENTS.md`, through the next heading of the same or higher level; `Graft` appends it; `Remove` cuts it; `Legacy` reports the `writrun:begin`/`writrun:end` markers a kit before `v0.0.04` left behind. `Replace` and the `yours` carry go with the fence they served.
6. Have `init` decide about `AGENTS.md` from the pointer rather than the markers, and take the project's `AGENTS.md` out of `update` entirely — neither read for writing nor written. The plan names a legacy fenced section as the adopter's to remove.
7. Replace `theGates` with a reading of `.writrun/gates.md`: every row of its table whose second cell is empty or still a TODO is one finding, named by the row's own first cell. `init`'s stage-1 gaps ask the same question of the same file.
8. Pin `v0.0.04`, refresh this repository with the resulting binary, and answer this repository's gates in `.writrun/gates.md` from the table its `AGENTS.md` carries today — moving answers, inventing none.
9. Cut this repository's fenced section from `AGENTS.md`, leave the pointer in its place, and replace the row routing a WritRun finding with the `routed` route `v0.0.04` states: the report is recorded here, the user says yes per finding, the issue opens upstream, and the local report ends `routed` naming it. `.ai/skills/docs/SKILL.md`'s law 0 restates that row and changes with it.

## Acceptance criteria (EARS)

- When a refresh runs, the system shall write every file the fetched template ships that no `Untouchable` path covers.
- When a refresh runs and the tag ships a file the previous one did not, the system shall write it without any Go change naming it.
- When a refresh runs, the system shall write every file it named in the plan.
- When a refresh runs and an adopter-owned file inside `.writrun/` is absent, the system shall write the template's copy.
- When a refresh runs and that file is present, the system shall leave it byte-for-byte unchanged.
- When a refresh runs, the system shall leave `AGENTS.md`, `CLAUDE.md`, `docs/` and `work/` byte-for-byte unchanged, `docs/writrun-instructions.md` excepted.
- When `AGENTS.md` carries a `writrun:begin` marker, the plan shall name it as the adopter's to remove, and the refresh shall proceed.
- When `init` runs where `AGENTS.md` already exists, the system shall append the `## WritRun` pointer section and change no byte above it.
- When `init` runs where `AGENTS.md` already links `.writrun/AGENTS.md`, the system shall leave the file alone.
- When `uninstall` runs, the system shall cut the pointer section, or a legacy fenced section, and leave every byte outside it.
- When a row of `.writrun/gates.md` is empty or still a TODO, `doctor` shall report one finding naming that row's transition.
- When `.writrun/gates.md` states a row this binary has never seen, `doctor` shall judge it by the same rule and name it by its own words.
- When `.writrun/gates.md` is absent, `doctor` shall report one finding naming the file.
- When `writrun doctor` runs against this repository after the refresh, the system shall report no stage-1 finding.
- When this repository's `AGENTS.md` is read after the refresh, it shall carry no `writrun:begin` marker and one link to `.writrun/AGENTS.md`.
- When `AGENTS.md` and `.ai/skills/docs/SKILL.md` are read together, neither shall forbid what the other allows about where a WritRun finding is filed.

## Edge cases

- A template shipping a file under `work/`: `Untouchable` covers it, so a refresh writes nothing there — the queue is the project's record whatever the tag says.
- A template shipping `docs/product/README.md` beside `docs/writrun-instructions.md`: `Untouchable` covers `docs`, `KitOwned` carves out the one file, and the project's chapters survive.
- A repository whose `AGENTS.md` carries both a legacy fenced section and the pointer: `uninstall` cuts the fenced section, which is the larger of the two and the one holding the kit's own prose.
- A repository adopted before `.writrun/gates.md` existed, whose `AGENTS.md` still holds the gates table: the refresh seeds the file with the template's TODOs, and `doctor` names each unanswered row. Nothing reads the old table, and nothing deletes it.
- `AGENTS.md` absent at update time: it is the project's file and a refresh has no opinion about it, so the refresh proceeds.
- A template whose `AGENTS.md` carries no link to `.writrun/AGENTS.md`: `init` stops rather than graft an empty section, as it stops today on a template carrying no `AGENTS.md`.
- `.writrun/gates.md` present with no markdown table: the file is unreadable as gates, which is one finding, not silence.

## Tests required

Go cases in `internal/pointer` over the four verbs, including a document
carrying both shapes and a document carrying neither.

Go cases in `updatecmd` for each acceptance criterion the refresh owns,
driven through the fake fetcher over a template built in the test: a
file no list names arriving, an adopter-owned file seeded once and never
again, `docs/` untouched but `docs/writrun-instructions.md` written,
`AGENTS.md` untouched, and the legacy fence named in the plan. The cases
asserting the fenced section is rewritten and the closed refresh list
are inverted, not deleted; the Outcome names each.

Go cases in `initcmd` for the graft, the already-pointed file, and the
template with no pointer.

Go cases in `doctorcmd` for a gates file that is absent, one that is all
TODO, one that is fully answered, one carrying a row this binary has no
knowledge of, and this repository's own.

An integration case driving `init`, then `update` from `v0.0.03` to
`v0.0.04` over a fixture repository, asserting the shape on disk.

## Definition of Done

- [ ] `writrun update` refreshes this repository from `v0.0.03` to `v0.0.04`.
- [ ] `.writrun/VERSION` records `v0.0.04` and `--version` names it.
- [ ] `.writrun/gates.md` holds this repository's answers, none of them a TODO.
- [ ] `writrun doctor` reports no stage-1 finding.
- [ ] A second `writrun update` reports only the recorded tag as differing.
- [ ] `AGENTS.md` carries the pointer and no fenced section, and states the `routed` route.
- [ ] No Go file names a kit file that is not `AGENTS.md`, `CLAUDE.md`, `.writrun/VERSION`, `.writrun/AGENTS.md`, `.writrun/gates.md`, `.writrun/settings.json`, `.writrun/conventions/` or `docs/writrun-instructions.md`.

## Proposed product changes

- `product/adoption/init.md` — what `init` does to an existing `AGENTS.md`.
- `product/adoption/update.md` — what a refresh rewrites, seeds and leaves alone.
- `product/adoption/uninstall.md` — which section of `AGENTS.md` is cut, and how the kit's files are recognised.
- `product/adoption/doctor.md` — where the gates are read and how they are judged.

## Proposed technical changes

- `technical/layout/tree.md` — `internal/fence/` becomes `internal/pointer/`.

## Outcome

_(fill after execution)_

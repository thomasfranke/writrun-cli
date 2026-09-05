---
id: spec-0004
task_ref: task-0004
status: implemented
created: 2026-09-03T22:30:37Z
---

# spec-0004 — Report every assumption's state, repair nothing

**References:** [task-0004](../tasks/task-0004-doctor-command.md)

- **Goal:** `writrun doctor` names every broken assumption; repairs none.

## Scope

In: checks grouped by stage, run from stage 0 up to the declared one.
Out: any write, any repair, any `--fix`; any check for a stage above
the declared one.

## Steps

1. Read the declared stage; run the groups from stage 0 up to it.
2. Stage 0 — environment: `git`, `bash` and POSIX `awk`/`sed` found on the `PATH`; each missing one named.
3. Stage 1 — files: an About file; at least one real product chapter; a technical doc; the `docs/` / `work/` split; the four gates the methodology requires answered in `AGENTS.md` (docs changes, rule declared finished, spec approval, a task without a spec); markers intact; `.writrun/VERSION` parseable; `check_front_matter.sh` exit 0; `check_settings.sh` exit 0.
4. Stage 2 — forge (via `gh api`): `gh` authenticated; workflow permissions read-and-write, so the recording bot can push to `main`; squash merging allowed; `main` reachable by the Actions bot — on the ruleset's bypass list where the forge offers one; the four rules that block the recording push named when on (`update`, `required_signatures`, `required_status_checks`, `pull_request` on user-owned repos).
5. Stage 3 — Issues enabled.
6. Recommendations reported as such; exit non-zero only when a finding breaks a flow; every finding names the file or setting and what is expected of it.

## Acceptance criteria (EARS)

- When every assumption holds, the system shall exit 0.
- When a script requirement is missing from the `PATH`, the system shall name it and exit non-zero.
- When a finding would break a flow, the system shall exit non-zero naming the file or setting and the expectation.
- When the forge is unreachable, the system shall report which checks it could not make and shall not fail them.
- When the declared stage is 1, the system shall make no forge read.
- When run, the system shall write nothing.

## Edge cases

- Stage 1 declared: no forge read is made; the stand-down is said.
- `gh` absent or unauthenticated: same as unreachable forge.

## Tests required

Integration with stubbed `gh`; one fixture per finding class.

## Definition of Done

- [x] Every check in scope has a passing and a failing fixture.
- [x] Suite green.

## Proposed product changes

- none — `product/adoption/doctor.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

`internal/command/doctorcmd/` carries the command, registered in
`commands()` between `update` and `uninstall`. It reads the world
through four ports and opens none of its own: `kit.Runner` for the
repository's scripts, `forge.Client.Run` for `gh`, `vfs.FS` for the
filesystem, `exec.LookPath` for the `PATH`.

The declared stage comes from the repository's own
`.writrun/scripts/stage-2-pull-requests/read_setting.sh`, so the
documented defaults are the reader's and not this binary's. A reader
that refuses stands the run down to stage 1 and reports that as a
finding.

Each finding carries a stage, a level and a sentence naming the file or
setting. Three levels: `breaks` sets the exit status, `advises` is a
recommendation, `unread` is a check the forge would not answer. Only
`breaks` reaches the exit status, so an unreachable forge and a missing
recommendation both exit 0. A wrapped check's own reporting is printed
under its finding unedited.

Two stage-2 answers are recommendations rather than breakages, against
the spec's plainest reading: `main` governed by no ruleset, and a
ruleset over `main` naming no bypass actor. Neither refuses a
fast-forward push on its own — `report-0010` records the evidence from
this repository. What refuses the push is one of the four rules, and
those are `breaks`.

`.writrun/VERSION` is parsed inside this package rather than through a
shared parser: `updatecmd` reads the same file to order two releases,
which is a different question. `report-0009` records the checks
`initcmd` and `doctorcmd` now write twice.

Tests: 84 Go cases in `internal/command/doctorcmd/` (97.5% of the
package's statements), and 49 assertions across nine case files in
`tests/integration/doctor/` on `tests/doctor_lib.sh`, where `gh` is a
stub answering from files named after its own arguments.

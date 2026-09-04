---
id: spec-0004
task_ref: task-0004
status: approved
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

- [ ] Every check in scope has a passing and a failing fixture.
- [ ] Suite green.

## Proposed product changes

- none — `product/adoption/doctor.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

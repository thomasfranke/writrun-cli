---
id: spec-0014
task_ref: task-0015
status: implemented
created: 2026-09-05T11:18:26Z
---

# spec-0014 — Every pull request is gated on coverage, races and known vulnerabilities

**References:** [task-0015](../tasks/task-0015-harden-the-go.md)

- **Goal:** the pipeline refuses what a reviewer would otherwise have to catch by hand.

## Scope

In: `.github/workflows/tests.yml`, `scripts/coverage.sh`, and the
`Makefile` targets that name them.

Out — already written: the technical docs state these rules already;
this change makes them true, and touches no permanent doc.

Out: the suite's own cases — this raises what CI reads, never what the
tests assert. Out: any gate on a branch other than a pull request's;
`main` keeps `release-readiness.yml`.

1. Raise the global coverage floor in `scripts/coverage.sh`'s default
   to the number `testing/tiers.md` states.
2. Add a **per-package floor**, checked from the same profile: a package
   below it fails the run naming itself and its percentage. A global
   average is what let two 0% packages ride green.
3. Run the unit tier with `-race`, `-shuffle=on` and an explicit
   `-timeout`.
4. Add `govulncheck` over the module graph as a step of its own, so a
   finding names itself rather than arriving as a failed build.
5. Add `staticcheck` after `go vet`, failing on its default checks.
6. Fail when `go mod tidy` would change `go.mod` or `go.sum`.
7. Pin every action by commit SHA, and add `concurrency` with
   `cancel-in-progress` so a superseded run stops.
8. Verify the pipeline against `testing/ci.md`'s order: every step it
   names, in that order, stopping at the first failure.

## Acceptance criteria (EARS)

- When total coverage over `internal/` is below 90%, the system shall
  fail the run naming the percentage and the floor.
- When any package under `internal/` is below the per-package floor, the
  system shall fail the run naming that package and its percentage.
- When a data race is detected, the system shall fail the run.
- When `govulncheck` reports a vulnerability the module graph reaches,
  the system shall fail the run naming the module and the advisory.
- When `go mod tidy` would change `go.mod` or `go.sum`, the system shall
  fail the run showing the diff.
- When a run is superseded by a newer push to the same pull request, the
  system shall cancel the older one.
- When the pipeline runs, the floors it enforces and the order it runs
  in shall be the ones `testing/tiers.md` and `testing/ci.md` state.

## Edge cases

- A package with no statements at all (a `const`-only package) divides
  by zero in a naive per-package check: it is not a failure and is
  reported as not applicable.
- `-race` needs cgo; the cross-compile step deliberately runs with
  `CGO_ENABLED=0`, so the two cannot share a step.
- `govulncheck` needs network. An outage must fail as an outage, named
  as such, never silently pass.

## Tests required

Integration over `scripts/coverage.sh`: a fixture profile above and
below the global floor, one with a single package below the per-package
floor, and one with a statement-less package — asserting the exit code
and that the failing package is named.

## Definition of Done

- [ ] Every criterion above has a passing and a failing fixture.
- [ ] The repository passes its own new gates at the merge.
- [ ] Suite green.

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- none — the rules were authored ahead of the work, in the pull request
  that derived this task. This change makes them true; a doc it also
  rewrote would be a rule approving itself.

## Outcome

`tests.yml` runs ci.md's nine steps in its order. Actions are pinned by
commit SHA; `staticcheck` and `govulncheck` are pinned by version tag and
run through `go run`, so neither reaches `go.mod`. `concurrency` is keyed
on the pull request.

`scripts/coverage.sh` gates the total and the per-package floor from one
profile. It sums statements from the profile rather than reading
`go tool cover -func`: `-coverpkg` writes one copy of every block per
test binary, so a block is counted once and is covered when any copy
covered it. `COVERAGE_GATE_ONLY=1` gates a profile without running a
test, which is how `tests/integration/coverage/` states the percentages
each case is about.

`staticcheck` reported four findings the repository cleared to pass its
own new gate. `ownedSkeletons` in `initcmd/plan.go` was dead — `plan`
keeps every file the project already has, not the two that map named.
`runInit` returned its error first. `vfs`'s `entry` built two `info`
struct literals a conversion covers. Reordering `runInit` exposed an
evaluation-order fault in the same helper: Go evaluates return
expressions left to right, so the printed output was read before `Run`
wrote it.

The repository passes both floors: 98.1% over `internal/`, lowest package
`internal/wrepo` at 90.9%.

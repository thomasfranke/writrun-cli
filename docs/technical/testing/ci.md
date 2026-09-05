# CI

Three workflows. `tests.yml` runs on pull requests;
`release-readiness.yml` re-verifies `main` on every push; `release.yml`
builds and publishes on a pushed `v*` tag. The `writrun-*` workflows
beside them are the kit's, not this project's.

## Order

`tests.yml` runs these in order and stops at the first failure. The
order is cheapest-first: a fault that `gofmt` can name costs seconds,
and the e2e tier costs minutes.

| | Step | Fails when |
|---|---|---|
| 1 | `gofmt -l` | any file would be rewritten |
| 2 | `go vet` | vet reports anything |
| 3 | `staticcheck` | a default check reports anything |
| 4 | `go mod tidy` | tidy would change `go.mod` or `go.sum` |
| 5 | cross-compile, `CGO_ENABLED=0` | any of darwin/arm64, linux/amd64, windows/amd64 fails to build |
| 6 | unit, with the coverage gate | a floor in [tiers](tiers.md#coverage) is not met |
| 7 | `govulncheck` | the module graph reaches a known vulnerability |
| 8 | integration | a case fails |
| 9 | e2e | a case fails |

The unit tier runs with `-race`, `-shuffle=on` and an explicit
`-timeout`. `-race` needs cgo, so it never shares a step with the
cross-compile above it.

## Rules

- **The coverage gate has one home**:
  [`scripts/coverage.sh`](../../../scripts/coverage.sh). CI calls it
  through `make cover` and a session runs the same script, so what the
  pipeline reads after a push is what a session reads before one
  ([makefile](../layout/makefile.md)).
- **A pull request is not approved below the floors.** The numbers are
  [tiers](tiers.md#coverage)'s; the gate reads them from the script.
- **Every action is pinned by commit SHA.** A moving tag is a
  dependency nobody reviewed.
- **A superseded run is cancelled**: `concurrency` is keyed on the pull
  request, `cancel-in-progress` on.
- **A network failure is named as one.** `govulncheck` needs the
  advisory database; an outage fails the step saying so, and never
  passes it quietly.

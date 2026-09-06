# Makefile

**Thin aliases only.** The scripts under
[`scripts/`](../../../scripts/) and [`tests/`](../../../tests/) are the
interface; the Makefile names them so a person does not have to. CI
calls these targets too, so renaming one is a workflow change.

| Target | Runs |
|---|---|
| `make tests` (alias `make test`) | every tier: `go test ./...`, then the bash suites |
| `make test-unit` | the Go unit tier alone, without the coverage gate |
| `make test-integration` | the compiled binary against fixture repositories |
| `make test-e2e` | a whole flow against a local WritRun clone |
| `make test-<suite>` | one suite directory, e.g. `make test-release`, `make test-update` |
| `make cover` | the unit tier with the coverage gate, printing the percentages and the floors |
| `make ui` | the screen against this repository — `writrun` with no command; needs a terminal |
| `make release [patch\|minor\|major]` | the whole release path; default `patch` |

- **`make cover` takes a floor**: `make cover 95` raises the total's
  floor for one run. Without an argument the floor is the one
  [`scripts/coverage.sh`](../../../scripts/coverage.sh) defaults to,
  which is the number [tiers](../testing/tiers.md#coverage) states.
- **`make tests` does not gate on coverage** — the gate is CI's, and
  `make cover` is how a session reads it before pushing.
- **`WRITRUN_BIN_DIR`** lets the integration cases share one compiled
  binary instead of relinking per case file; the targets set it.

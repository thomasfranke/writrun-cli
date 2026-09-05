# Tiers

| Tier | Proves | Lives in | How |
|---|---|---|---|
| unit | one package's logic, every port faked | `*_test.go` beside the code | `go test ./...`, table-driven |
| integration | the compiled binary against fixture repositories, `gh` and the agent stubbed | `tests/integration/` | one directory per subject, one file per behaviour |
| e2e | a whole flow — adopt, take, finish — against a local WritRun clone and a bare origin | `tests/e2e/` | the forge is the one fake |

## Coverage

- **Coverage gates the pipeline**, over `internal/`, on two floors: the
  total is at least **90%**, and no single package is below **80%**.
- **The per-package floor is what the total cannot see.** A whole
  package at 0% rides a high average, and two of them did — `updatecmd`
  and `uninstallcmd` reached the forge uncovered under a total-only
  gate.
- A package holding no statements is reported as not applicable, not as
  a failure.
- **The gate is enforced in CI and readable locally**: `make cover`
  prints the percentages and the floors, running the same script CI
  runs ([ci](ci.md#rules)). `make tests` does not gate.

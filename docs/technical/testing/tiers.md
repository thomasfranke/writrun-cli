# Tiers

| Tier | Proves | Lives in | How |
|---|---|---|---|
| unit | one package's logic, every port faked | `*_test.go` beside the code | `go test ./...`, table-driven |
| integration | the compiled binary against fixture repositories, `gh` and the agent stubbed | `tests/integration/` | one directory per subject, one file per behaviour |
| e2e | a whole flow — adopt, take, finish — against a local WritRun clone and a bare origin | `tests/e2e/` | the forge is the one fake |

- **Coverage gates the pipeline**: `go test -cover` over `internal/`
  fails CI below 85%. The gate lives in CI, never in a local run.

# Tiers

| Tier | Proves | Lives in | How |
|---|---|---|---|
| unit | one package's logic, every port faked | `*_test.go` beside the code | `go test ./...`, table-driven |
| integration | the compiled binary against fixture repositories, `gh` and the agent stubbed | `tests/integration/` | one directory per subject, one file per behaviour |
| e2e | a whole flow — adopt, take, finish — against a local WritRun clone and a bare origin | `tests/e2e/` | the forge is the one fake |

[`tests/README.md`](../../../tests/README.md) is the suite's own index:
where each tier is on disk, why the unit tier is not under `tests/`, and
what each tier can prove that the others cannot.

## Signals

- **A unit case may raise a real signal at the test binary only where
  the package holds a registration no case owns.** A Go process takes
  the disposition that kills once nothing is registered for a signal,
  and both `signal.Stop` on the last channel and `signal.Reset` put it
  back there.
- **Delivery is asynchronous**, so a signal raised before a teardown can
  arrive after it — which is why the registration a case makes for
  itself is not enough.
- **A binary that dies of its own fixture reports the package `FAIL`
  with no case named**, on whichever pull request was running.
- `internal/command/finishcmd` holds that registration in `TestMain` and
  re-arms it wherever a case resets a signal.
- **The cases keep raising real signals.** What they prove is that a
  signal arriving between the completion writes and the confirmation
  puts the writes back, and a faked signal proves a faked window.

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

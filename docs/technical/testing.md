# Testing

- `make tests` runs [`tests/run.sh`](../../tests/run.sh): one tier per
  directory (`integration/`, `e2e/`), one directory per script under
  test, one file per behaviour.
- Every case sources its domain fixture (`tests/release_lib.sh`,
  layered on `tests/harness.sh`) and also runs standalone.
- CI runs the suite on pull requests (`tests.yml`) and re-verifies main
  on every push (`release-readiness.yml`).
- Makefile aliases: `make tests` (alias `make test`) for the whole
  suite, `make test-unit` / `make test-integration` for a tier,
  `make test-<suite>` (e.g. `make test-release`) for one suite
  directory.

# Suites

- The bash suites run on
  [`tests/harness.sh`](../../../tests/harness.sh): the release suite
  under `tests/*/release/`, the CLI integration cases under
  `tests/integration/cli/`.
- Makefile aliases: `make tests` (alias `make test`) for every tier,
  `make test-unit` / `make test-integration` / `make test-e2e` for one,
  `make test-<suite>` (e.g. `make test-release`) for one suite
  directory.

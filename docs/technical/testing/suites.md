# Suites

**No file lists the suites.** [`tests/run.sh`](../../../tests/run.sh)
walks `tests/` — one directory per tier, one directory per subject
inside it, one `*_test.sh` file per behaviour — so a case is registered
by its path and nowhere else. `make test-<suite>` resolves the same way,
by glob, and exits 3 naming the suite when nothing matches.

Every case sources the fixture for its domain. Every case also runs
standalone:

```bash
bash tests/integration/release/minor_bumps_middle_digit_test.sh
```

## Fixtures

Each layers on the one before it, so a case sources exactly one.

| Fixture | Layered on | Sourced by |
|---|---|---|
| [`harness.sh`](../../../tests/harness.sh) | — | `tests/e2e/release/` |
| [`cli_lib.sh`](../../../tests/cli_lib.sh) | `harness.sh` | `tests/e2e/adopt/`, `tests/integration/cli/` |
| [`coverage_lib.sh`](../../../tests/coverage_lib.sh) | `harness.sh` | `tests/integration/coverage/` |
| [`release_lib.sh`](../../../tests/release_lib.sh) | `harness.sh` | `tests/integration/release/` |
| [`init_lib.sh`](../../../tests/init_lib.sh) | `cli_lib.sh` | `tests/integration/init/`, `tests/integration/uninstall/`, `tests/integration/update/` |

The Makefile's targets are [Makefile](../layout/makefile.md)'s.

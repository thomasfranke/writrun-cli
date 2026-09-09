# The suite

Three tiers, and this is where to find all of them. Two live here; the
third cannot, and the reason is worth knowing before looking for it.

| tier | what it drives | where | how to run |
| --- | --- | --- | --- |
| unit | one package's logic, every port faked | `*_test.go` **beside the code**, in `internal/` and `cmd/` | `make test-unit` |
| integration | the compiled binary against fixture repositories | `tests/integration/` | `make test-integration` |
| e2e | a whole flow against a real WritRun clone and a bare origin | `tests/e2e/` | `make test-e2e` |

`make tests` runs all three. `make cover` runs the unit tier with the
coverage floors, which gate the pipeline
([tiers.md](../docs/technical/testing/tiers.md)).

## Why the unit tier is not in this directory

Go decides this, not taste. A test that reaches a package's unexported
names must be compiled *into that package*, and a Go package is a
directory — so those files have to sit next to the code they test.
Seventy-eight of this repository's eighty-one Go test files are of that
kind: they reach `newSession`, `waitForLine`, `dispatch`, `newEntry`.

Moving them under `tests/unit/` would mean exporting everything they
touch, which opens the internal API that
[the coupling guard](../docs/technical/engineering/coupling.md) exists
to keep shut. The cost is real and one-directional, so the tests stay
where the language puts them and this table says where that is.

## A case never reads the terminal that ran the suite

Every case is given `/dev/null` on stdin, by `tests/run.sh` and by the
Makefile's own loops. A case that needs a terminal stands one up itself
through `WRITRUN_TTY_IN`; a case checking what a command does *without*
one is then asserting about a stdin the suite owns, rather than about
whoever happened to run it.

This is not tidiness. Inherited from a keyboard, such a case asks its
question for real and waits — and inside `make release`, where the
suite runs nested and its output is captured, it waits somewhere nobody
can see. The release neither ends nor says why
([report-0031](../work/reports/report-0031-suite-inherits-the-terminal.md)).

## The shape of a case here

One directory per subject under test, one file per behaviour, suffixed
`_test.sh`. Every case sources the fixture for its domain, layered on
`harness.sh`, and also runs standalone:

    bash tests/integration/release/minor_bumps_middle_digit_test.sh

`tests/run.sh` discovers every tier and every case; it is what `make
tests` calls.

## What each tier can prove, and what it cannot

The tiers are not the same test at three sizes — each sees something
the others cannot.

- **unit** drives a model: keys in, actions out. It cannot see a
  terminal, so a model that answers correctly while the terminal under
  it does not is invisible here.
- **integration** drives the compiled binary against fixtures. It sees
  the binary's behaviour, still without a terminal.
- **e2e** drives whole flows, and `tests/e2e/screen/` drives a real
  terminal through `expect`. It is the only tier that can watch a
  screen: three UI defects reached a person before it existed, and none
  of them was visible to a case that never opened a pty.

An assertion belongs in the lowest tier that can actually make it. Some
cannot go low: huh writes nothing to an output that is not a terminal,
so what a question *draws* is only assertable in the e2e tier — while
what it *answers* belongs in `internal/term`.

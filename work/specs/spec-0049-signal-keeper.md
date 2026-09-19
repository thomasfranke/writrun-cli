---
id: spec-0049
task_ref: task-0041
status: draft
created: 2026-09-19T23:36:42Z
---

# spec-0049 — The package holds the registration no case may drop

**References:** [task-0041](../tasks/task-0041-signal-keeper.md)

- **Goal:** a signal a case raises can never kill the test binary, whatever the case did with its own registration, and the rule that makes that true is written where the tiers are stated.

## Scope

In: `internal/command/finishcmd/signals_test.go` and
`helpers_test.go` — a `TestMain` for the package, `unignore`, and one
new case; `technical/testing/tiers.md`, which gains the rule.

Out: `internal/command/finishcmd/signals.go` and everything else the
guard does. The production code is not what failed; the fixture is.
Out: the other packages, none of which raises a signal.

**The load-bearing premise:** a Go process kills itself on SIGTERM only
where no channel is registered for it. `signal.Stop` on the last
channel and `signal.Reset` both restore that state, and delivery is
asynchronous, so a signal raised before either can arrive after it. If
the runtime kept its handler installed regardless, the observed death
has another cause and this spec's shape is wrong.

## What is wrong

`raise()` sends the signal to `os.Getpid()`. Every case that calls it
registers something first — the guard under test, or a probe — and
every one of those registrations ends inside the case: the guard's when
`run` returns, a probe's on `defer signal.Stop(probe)`, and both
`signal.Ignore` cases undo themselves through `unignore`, whose
`signal.Reset` is exactly the call that restores the killing
disposition.

The window is small and it is real: run
[35474055199](https://github.com/thomasfranke/writrun-cli/actions/runs/35474055199)
died at 0.554s, against a 10-minute timeout, on a pull request that
changed no code this package reads. Eight local runs at the seed that
job printed all passed, so the trigger is timing and not order.

## Steps

1. Give the package a `TestMain` that registers a keeper channel for `caught` before any case runs, drains it for the life of the run, and never stops it. A second registration is what makes an individual `signal.Stop` a no-op for the disposition.
2. Re-arm the keeper inside `unignore`, after its `signal.Reset` — `Reset` undoes every registration for the signal, the keeper's included, and a keeper that two cases can drop is not a keeper.
3. Add the case that holds the invariant: raise SIGTERM with nothing of the case's own registered, and assert the binary is still running afterwards. Without the keeper the process dies and the run says so.
4. State the rule in `technical/testing/tiers.md`: a unit case may raise a real signal only where the package holds a registration no case owns, and say why the fixture may not fake the signal.

## Acceptance criteria (EARS)

- When a case raises SIGTERM with no registration of its own, the system shall continue running and the case shall pass.
- When a case stops its own probe and a raised signal arrives after the stop, the system shall not terminate.
- When `unignore` returns, the system shall still hold the keeper's registration for every signal it reset.
- When the whole package is run with `-race -shuffle=on` one hundred times, the system shall not report `signal: terminated`.

## Edge cases

- A binary started with SIGINT ignored — `writrun finish &` in a shell, and the reason `TestAnInterruptInTheWindowPutsTheWritesBack` skips: `signal.Notify` does not un-ignore a signal, so the keeper changes nothing there and the skip stays.
- `TestAGuardWithNothingToArmForRegistersNothing` ignores both signals and asserts the guard registers nothing: the keeper is a registration by the fixture, not by the guard, and the case reads the guard's own channels.
- A case that asserts the guard died of a signal still reads its own channel; the keeper only drains what nobody else took.
- The keeper must not swallow what a probe is waiting for: `signal.Notify` delivers to every registered channel, so both receive.

## Tests required

The new case above, and it is the one held against its absence: with
the `TestMain` keeper removed, it kills the test binary and the package
reports `signal: terminated` — the failure this change exists to
prevent, reproduced on demand rather than waited for.

The package's existing signal cases must keep passing unchanged: they
are the proof that the guard answers a real signal, and a change that
made them stop raising one would have removed the coverage instead of
the flake.

## Definition of Done

- [ ] A raise with no case-level registration does not kill the binary.
- [ ] `unignore` leaves the keeper armed.
- [ ] The existing signal cases pass unchanged.
- [ ] `technical/testing/tiers.md` states the rule.

## Proposed product changes

- none — nothing a user of the binary can observe changes.

## Proposed technical changes

- `technical/testing/tiers.md` — the rule a unit case raising a real signal obeys.

## Outcome

_(fill after execution)_

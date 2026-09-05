---
id: spec-0016
task_ref: task-0017
status: implemented
created: 2026-09-05T13:58:36Z
---

# spec-0016 — The fetch is faked at its own boundary, not underneath it

**References:** [task-0017](../tasks/task-0017-give-the-fetch.md)

- **Goal:** a test asks the fetch for a kit and gets one, without a clone and without a network.

## Scope

In: `kitfetch` behind an interface its consumers name, with a fake
beside it; `initcmd` and `updatecmd` taking that interface; the
production wiring in `cmd/writrun/` alone.

Out — **the fetch's own behaviour**. What it does with `git clone`, the
shallow depth, the tag it asks for and the refusal when the clone
carries no `template/` all stay exactly as they are. This gives the
boundary a fake; it does not move the boundary.

Out: any change to `internal/vfs`. The filesystem port is not what is
missing here, and a fake filesystem underneath a real `git clone` is
what this task exists to stop attempting.

1. Name the interface where it is consumed: one method that takes a tag
   and a source and hands back a template directory and its cleanup —
   `kitfetch.Fetch`'s signature is already that shape.
2. Write the fake beside it: a template tree it hands over, and a way to
   say *this tag fails*, so a test names the failure rather than
   arranging a clone that cannot work.
3. Move `initcmd` and `updatecmd` onto it.
4. Write the unit tests the seam was blocking — the partial-state
   messages of both commands, which name `git checkout -- .` and
   `git clean -fd`, and which no fixture reaches today.
5. Keep one case per command against the real fetch. A fake that has
   never been checked against the thing it fakes is a second
   implementation nobody compares.

## Acceptance criteria (EARS)

- When a unit test drives `init` or `update`, the system shall make no
  clone and reach no network.
- When the fake is told a tag fails, the command shall report the
  failure naming the tag and the source, and shall write nothing.
- When either command's write fails after the fetch succeeded, the
  system shall report the partial state and name the two git commands
  that undo it — proven by a test, which today it is not.
- When the suite runs, at least one case per command shall exercise the
  real fetch against a local repository.
- When the refactor is complete, every case in `tests/integration/`
  shall pass with no edit to the case.

## Edge cases

- The cleanup is part of the contract: a fake that hands back a
  directory and no cleanup lets a test pass while the real one leaks.
- `kitfetch` verifies the clone carries `template/`. That check belongs
  to the real implementation, so the fake must be able to answer "a
  repository, but not a WritRun one" without a clone.

## Tests required

Unit, per command: the fake handing over a template, and the fake
refusing a tag. Integration: unchanged. One case per command against a
local repository, so the fake and the real fetch are compared rather
than assumed equal.

## Definition of Done

- [x] No unit test clones.
- [x] The partial-state message of each command is proven by a test.
- [x] One case per command still exercises the real fetch.
- [x] The integration suite passes unmodified.
- [x] Suite green.

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- `technical/layout/tree.md` — the row for `internal/kitfetch/` says
  what it is; it will need to say that it carries a fake too.

## Outcome

`kitfetch.Fetcher` is the interface: `Fetch(tag, source)` handing back
a template directory and its cleanup. `kitfetch.Clone` is the
production implementation and `kitfetch.Fake` the one the tests
inject; both refusals are built by the constructors the real fetch
uses, so the fake says what the user is told. `Fail` answers a clone
that could not run, `FailNoTemplate` a repository that is not a
WritRun one, and the handed-over cleanup is counted, so a leak is
assertable. `initcmd.Deps.Kit` and `updatecmd.Deps.Kit` carry it;
`cmd/writrun/main.go` is the only production wiring that changed.

The clone's own behaviour is untouched: the depth, the tag it asks for
and the refusal when the clone carries no `template/` are exactly as
they were. `internal/vfs` is untouched.

Both partial-state messages are now proven — `git checkout -- .` and
`git clean -fd` for each command, over a fake filesystem refusing the
write the fetch had already been paid for. Two further cases assert
the cleanup runs whichever way the command ended. No unit case of
either command clones; `TestInitAdoptsEndToEnd` and
`TestRefreshMovesTheKitAndLeavesTheProject` keep the real fetch
against a local repository.

Go suite: 274 cases green in 15 packages.
`make test-integration`: 104 assertions, 0 failures, with no edit to
any case. `bash tests/run.sh`: 48 case files passed, 0 failed.
`make cover`: 98.3% over `internal/` (floor 90%), `internal/kitfetch`
at 100% (floor 80%).

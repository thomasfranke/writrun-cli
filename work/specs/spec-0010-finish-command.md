---
id: spec-0010
task_ref: task-0011
status: implemented
created: 2026-09-03T22:30:42Z
---

# spec-0010 — Verify promises, record outcome, mark ready

**References:** [task-0011](../tasks/task-0011-finish-command.md)

- **Goal:** `writrun finish` closes the loop in the load-bearing order.

## Scope

In: the completion sequence — deltas verified, outcome recorded, spec
`implemented`, `completed` written, state checked, PR marked ready. Out:
the task's `status` line (the machinery's, always); the merge.

## Steps

1. Run `check_deltas.sh` for the branch's specs; stop on non-zero — nothing else runs.
2. Require the spec's `## Outcome` filled; write `status: implemented` on the spec and the task's `completed` UTC timestamp. Touch the task's `status` line never.
3. Run `record_provenance.sh` unconditionally — it reads the ledger setting itself.
4. Run `preflight.sh`; exit 0 required.
5. Show what will happen; on confirmation mark the pull request ready for review.

## Acceptance criteria (EARS)

- When the delta check fails, the system shall stop before writing any status.
- When the spec's Outcome is empty, the system shall refuse to mark it implemented.
- When finish completes, the task's `status` line shall be unchanged.
- When preflight is non-zero, the pull request shall not be marked ready.

## Edge cases

- A task with empty `spec_ref`: no deltas to check; `completed` is still written; preflight still gates.
- Several specs on one branch: one `check_deltas.sh` call carrying all of them.

## Tests required

Integration over a fixture adoption: green path, each failing gate,
multi-spec branch.

## Definition of Done

- [x] The order of operations matches `finishing.md`'s and is asserted by tests.
- [x] Suite green.

## Proposed product changes

- none — `product/pull-requests/shape.md` already states the shape.

## Proposed technical changes

- none.

## Outcome

`internal/command/finishcmd/` runs the five steps in the fixed order and
reimplements none of them: `check_deltas.sh` for the branch's specs in
one call carrying all of them, the `## Outcome` requirement, the spec's
`status: implemented` and the task's `completed` stamp,
`record_provenance.sh` unconditionally, `preflight.sh`, and — only then —
the composition shown and `gh pr ready` on a yes. Every helper it needed
stayed inside the package; `internal/kit`, `internal/forge` and the
frame's `Ctx`/`Terminal` are untouched, so `author` and `amend` inherit
a shape rather than a modified frame. Coverage: 93.6% of the package,
97.6% over `internal/`; 40 Go tests in the package and 36 assertions
across seven integration cases on the real kit scripts under
`tests/integration/finish/`.

Four things the steps did not say, decided here and left visible:
**the task** is the argument or, absent one, the branch's — the same
`task/NNNN-` inference `preflight.sh` makes; **the range** is
`origin/main...HEAD`, falling back to the local `main` and overridable
with `--range`, again preflight's own resolution; **the ledger's
vocabulary** (`--by`, `--login`, `--model` and the four counts) is
passed through unread, because `record_provenance.sh` validates it and a
second opinion would be a second authority; and **the two writes are
idempotent** — a spec already `implemented` and a `completed` date
already declared are reported, never restamped.

Two things noticed and recorded rather than settled: the completion
edits are never committed, so preflight's two range-reading stages judge
a branch that does not yet carry them (report-0014); and a declined
finish exits with those edits already made, which `shape.md`'s "a
refused command leaves nothing behind" does not describe
(report-0015). `finishing.md` is upstream WritRun's flow document and is
not in this repository — the order asserted by
`TestTheGreenPathRunsTheSequenceInOrder` and by the integration cases is
the one this spec's **Steps** fix.

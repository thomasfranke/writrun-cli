---
id: spec-0013
task_ref: task-0014
status: implemented
created: 2026-09-03T23:27:24Z
---

# spec-0013 — Read the branch, the checks and the queue; write nothing

**References:** [task-0014](../tasks/task-0014-status-command.md)

- **Goal:** `writrun status` reads where the work stands; it writes nothing.

## Scope

In: resolving the branch to a task, the read-only run of the completion
checks, the open-report count, the kit-tag comparison. Out: any write;
any repair; any judgement the checks do not already make.

## Steps

1. Resolve the current branch to a task by the `task/NNNN-*` name; no task — say so and skip step 3.
2. Name the task's spec and the spec's status from the queue files.
3. Run the completion checks (`preflight.sh`) read-only; name the first that fails — what `finish` would stop at — or that all pass.
4. Count the reports with status `open` in `work/reports/`.
5. Compare `.writrun/VERSION` with the tag this client pins; name a mismatch, bridge nothing.

## Acceptance criteria (EARS)

- When run, the system shall write nothing.
- When the branch carries no task, the system shall say so and still report the open reports and the kit tag.
- When a completion check fails, the system shall name that check and its failure.
- When `.writrun/VERSION` differs from the pinned tag, the system shall name both values.

## Edge cases

- Detached HEAD, or `main`: the no-task path, said plainly.
- A branch named like a task the queue does not hold: named as unknown, never invented.

## Tests required

Integration over fixture repositories: with and without a task branch,
passing and failing checks, matching and mismatched kit tags.

## Definition of Done

- [x] Every acceptance criterion has a test.
- [x] Suite green.

## Proposed product changes

- none — `product/queue/status.md` already states the behaviour.

## Proposed technical changes

- none.

## Outcome

`writrun status` ships as `internal/command/statuscmd/`, registered in
`commands()`. It answers in six labelled lines — branch, task, spec,
checks, reports, kit — and it exits 0 having answered: a failing check
is the answer, not a failure of the command, the reading `list` already
makes of "nothing is available".

The five steps landed as specified. Step 3 runs the repository's own
`preflight.sh` through the `kit` port with no arguments, so the task and
the range are inferred exactly as a completion run would infer them; the
verdict is preflight's own `PREFLIGHT STOPPED at n/3 <stage> — exit c`
line, quoted rather than summarised, because the script already names
the first stage that fails and a second reading of the queue would be a
second authority. Its own refusals — the ones that name no stage — are
quoted with the exit code beside them.

Two decisions worth recording. The step-1 resolution distinguishes *no
task* from *a task the queue does not hold*: both skip the checks, and
the second names the id as unknown rather than inventing a file for it.
And `.writrun/VERSION` is read, and the two tags compared, inside this
package — `updatecmd` reads the same file to decide a refresh and
`doctor` (task-0004) will read it to report on it, so the duplication
was taken deliberately over a shared package three commands in flight
would have to agree on before any could land; report-0008 records that
choice for triage.

Tests: 27 Go cases in the package (96.9% of its statements, over the
80% floor) and eight
bash case files under `tests/integration/status/`, on fixture
repositories carrying this repository's own kit — the completion checks
are copied, never stubbed, so the failing-check case fails through the
real front-matter sweep. Every acceptance criterion has a case, the
write-nothing one across all three paths.

No doc changed, as promised: `product/queue/status.md` already stated
the behaviour, and it needed no correction.

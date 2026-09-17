---
id: spec-0036
task_ref: task-0033
status: implemented
created: 2026-09-13T22:08:01Z
---

# spec-0036 — Name every requirement, mark it, and preview the next stage

**References:** [task-0033](../tasks/task-0033-stage-requirements.md)

- **Goal:** `doctor` names every requirement each stage makes, marks it met or not, and previews the stage above the declared one.

## Scope

In: the report's shape — a row per requirement, a count per stage,
four marks in place of the level words — and the preview of the stage
above the declaration.

Out: the screen and the explanations it shows ([spec-0037](spec-0037-doctor-screen.md));
any repair; any new check. The checks are the ones the binary already
makes.

## Steps

1. Give every check a requirement with a stable name, so a check that
   holds is a row rather than silence. The names are the checks' own:
   `internal/requirements` for stage 0, `doctorcmd/files.go` for stage 1,
   `doctorcmd/forge.go` for stages 2 and 3.
2. Render each stage as a heading that counts its rows — `Stage 1 —
   files: 8 of 9 met.` — with one row per requirement under it.
3. Mark each row `✓` met, `✗` breaks, `!` advises, `?` unread, and drop
   the level words. The glyph carries the state; the colour repeats it
   and carries nothing alone ([rules.md](../../docs/product/rules.md)).
4. Examine stages 0 to the declared one and preview the one above it,
   through the code path `--at` already uses. No check is written twice
   for the preview.
5. Keep the exit status reading the declared stages alone, and end the
   report with whether the next stage is within reach.
6. A read the preview needs and cannot make is `unread`. At stage 1 the
   preview is the first forge read the run makes, and a forge that does
   not answer leaves the requirements unknown, never unmet.

## Acceptance criteria (EARS)

- When every requirement holds, the system shall print a row for each
  one and exit 0.
- When a requirement is unmet, the system shall mark its row `✗` and
  name the file or setting and what is expected of it.
- When the declared stage is below 3, the system shall preview the stage
  above it and count what it would require.
- When a previewed requirement is unmet, the system shall leave the exit
  status unchanged.
- When the forge does not answer a previewed check, the system shall
  mark it `?` and exit 0.
- When stage 3 is declared, the system shall preview nothing and say so.
- When `--at` names a stage, the examined range shall follow it and the
  exit status shall still answer for the declaration alone.
- When colour is off, the four states shall still be distinguishable.

## Edge cases

- **Stage 3 declared.** There is no rung above; the report says so
  rather than inventing one.
- **No `gh`, no network.** Every previewed forge requirement is `?`, and
  the run exits 0.
- **A check a newer kit adds.** It has no name here, so its own sentence
  is the row — a requirement this binary has never heard of is still
  reported ([coupling.md](../../docs/technical/engineering/coupling.md)).

## Tests required

Unit, `internal/command/doctorcmd`: a met check renders a row; the
counts match the rows; the preview range follows the declaration; the
exit status ignores the preview; the marks differ without colour.

Integration, `tests/integration/doctor/`: one fixture per state, and a
stage-1 fixture whose stage-2 preview is entirely `unread`.

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`adoption/doctor.excalidraw`](../../docs/product/screens/adoption/doctor.excalidraw)
  — `writrun doctor — stage 2 within reach, exit 0` and `writrun doctor
  — stage 2 out of reach, exit 0`.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [ ] Every check names a requirement, and every requirement is a row.
- [ ] The preview runs from the declaration, and never reaches the exit
      status.
- [ ] Passing and failing fixtures for each mark.

## Proposed product changes

- `product/adoption/doctor.md` — the report's shape, the four marks,
  the preview and what it never affects.

## Proposed technical changes

- none — no machinery change.

## Outcome

Every check now answers a `requirement` — a stage, a stable name, one
of four marks, what the row adds after the name, and a wrapped
script's own words under it. `internal/command/doctorcmd/report.go`
renders them: a heading per stage counting its rows, `✓ ✗ ! ?` in place
of the level words, the rung above the declaration previewed, and two
closing lines — what the declaration answers for, and whether the next
stage is within reach. The preview runs `examine` at `declared + 1`,
the same call `--at` makes, so no check is written twice; the exit
status still reads `upTo(declared)` alone.

Both frames of `adoption/doctor.excalidraw` are asserted line for line
against what the binary renders
(`internal/command/doctorcmd/report_frames_test.go`). The harness reads
the `.excalidraw` JSON and reconstructs a captioned frame's lines from
each text element's position — there was none before, and one was
built.

**What the plan did not foresee.**

- **Four things the drawing settles that the spec did not state**: the
  report's lines carry a one-space inset, rows sit at three, a
  wrapped script's words at eight, and the pipe the old report drew
  before them is gone. All four are the frames', and the binary came to
  them.
- **A tenth stage-1 row exists, conditionally.** Reading the declared
  stage is a check like any other, but the frames draw nine rows. It is
  reported as a `writrun/settings.json` requirement only where the read
  fails, so a healthy repository counts nine and a broken one counts
  ten. The alternative — always a row — would have put the binary a row
  out from its own drawing.
- **`mainReachable` became two requirements**, `main is governed by a
  ruleset` and `no rule over main refuses the recording push`, because
  the frames draw two rows for one read. Where that read fails the
  reason sits under the second of them, once: printing it twice would
  read as two faults.
- **The frames wrap a long detail to fit the canvas** and the binary
  writes it whole. The case unwraps both sides, which is the one
  difference between them that is the canvas's.
- **At stage 1 the preview is now the first forge read a run makes.**
  `stage_one_reaches_no_forge_test.sh` asserted the opposite and was
  rewritten: what it guards now is that the rung above the preview is
  not reached.

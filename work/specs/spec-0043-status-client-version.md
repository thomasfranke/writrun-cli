---
id: spec-0043
task_ref: task-0035
status: implemented
created: 2026-09-13T22:29:38Z
---

# spec-0043 — status names the client as well as the kit

**References:** [task-0035](../tasks/task-0035-binary-explains.md)

- **Goal:** `status` names the client that answered as well as the kit installed.

## Scope

In: one row.

Out: everything else `status` prints, which this spec leaves alone.

## Steps

1. Print a `Client` row naming the product and its version, above the
   `Kit` row.
2. Read it where `--version` reads it, so the two surfaces cannot
   disagree about what is running.
3. Leave `Kit` exactly as it is: the tag recorded in this repository,
   and the mismatch with the pinned one that it already names.

## Acceptance criteria (EARS)

- When `status` runs, the system shall print the client's product and
  version.
- When the installed kit's tag differs from the pinned one, the system
  shall still name that difference.
- When the repository was never adopted, the system shall refuse as it
  does today.

## Edge cases

- **A development build.** The version prints as the binary reports it,
  `+dirty` included: a build that is not a release should not be able to
  look like one.

## Tests required

Unit, `internal/command/statuscmd`: the row's presence and its source.

Integration, `tests/integration/status/`: a run naming both the client
and the kit.

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`tasks/status.excalidraw`](../../docs/product/screens/tasks/status.excalidraw)
  — `writrun status — the rows explained, and the client named`.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [x] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [x] Both rows print, from the sources `--version` and the kit use.
- [x] [report-0034](../reports/report-0034-status-client-version.md)
      closed.

## Proposed product changes

- `product/queue/status.md` — the row and what it answers.

## Proposed technical changes

- none — no machinery change.

## Outcome

`status` prints `Client   writrun-cli <version>` above `Kit`. The
version is `command.Ctx.Version`, which the frame fills from the field
`--version` prints — one value, so the two surfaces cannot answer
differently about one process. `Kit` is untouched, and a build from
source still says `dev`.

The unit case compares the whole answer with the frame
`writrun status — the rows explained, and the client named`, read out
of the drawing at test time by `internal/drawing`. The fixture takes
the client and the tag from that frame's own rows rather than spelling
either: a frame states the facts it was drawn against, and a fixture
with its own copy would be a second opinion about what the rows say.
The integration case compares the printed row with what `--version`
answered in the same run.

Two things the plan did not foresee.

The frame's window is 70 columns, and the drawn `Task` row stops short
of the whole title: the fixture's task is called `Answer where the work
stands from the current branch`, and the row carrying it is 86
characters. The comparison is therefore per row — the frame's row is
what the printed one opens with — and the row count must match
exactly.

The drawing's window carries a paragraph under the rows, below the
frame's divider, explaining `Checks`, `Client` and `Kit`. It reads as
the drawing's own annotation rather than output: the `--version` frame
in `help.excalidraw` carries the same kind of grey block under the one
line that command prints. Nothing here prints it, and this spec's scope
— one row — is what was built.

The Proposed-changes entry was rewritten from a markdown link to the
bare backticked path that both gates read
([report-0041](../reports/report-0041-invisible-promises.md)).

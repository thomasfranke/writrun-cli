---
id: spec-0043
task_ref: task-0035
status: approved
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

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [ ] Both rows print, from the sources `--version` and the kit use.
- [ ] [report-0034](../reports/report-0034-status-client-version.md)
      closed.

## Proposed product changes

- `product/queue/status.md` — the
  row and what it answers.

## Proposed technical changes

- none — no machinery change.

## Outcome

_(fill after execution)_

---
id: spec-0037
task_ref: task-0033
status: approved
created: 2026-09-13T22:08:03Z
---

# spec-0037 — Make the report navigable, with the document's own sentences

**References:** [task-0033](../tasks/task-0033-stage-requirements.md)

- **Goal:** A reader moves over the requirements and reads, in the document's own words, what each one is and what would clear it.

## Scope

In: selectable requirement rows, the footer that names the selected
one, and the table in `doctor.md` the binary embeds to fill it.

Out: the checks and the report's shape ([spec-0036](spec-0036-doctor-requirements.md));
any repair; any write of any kind.

## Steps

1. Make every requirement row selectable — the met ones included — and
   show the headings, counts and blank lines while skipping them, as
   `internal/screen`'s `Row.Selectable` already does for the queue.
2. Render the footer in the shape
   [`adoption/config.excalidraw`](../../docs/product/screens/adoption/config.excalidraw)
   draws: a rule, `requirement — what it is` with a hanging indent, then
   the keys.
3. Give [`adoption/doctor.md`](../../docs/product/adoption/doctor.md) a
   table with one row per requirement: its name, what it is, why the
   stage needs it, what clears it.
4. Embed that table at build time and look the row up by requirement
   name, so the screen and the document cannot disagree.
5. Where the table has no row for a requirement, print the check's own
   sentence alone. A requirement a newer kit adds is explained by
   whatever the check says, and never by a guess.
6. `r` runs the checks again and keeps the cursor where it was.

## Acceptance criteria (EARS)

- When a requirement is selected, the system shall name it and print
  its explanation under the list.
- When a met requirement is selected, the system shall explain it too.
- When the table holds no row for the selected requirement, the system
  shall print the check's own sentence and nothing else.
- When the screen is open, the system shall write nothing.
- When `r` is pressed, the system shall re-run the checks and keep the
  selection.
- When the terminal is too short for the footer, the list shall scroll
  and the footer shall stay.

## Edge cases

- **A sentence longer than the pane.** It wraps under the hanging
  indent; it is never truncated to fit.
- **Re-running with the forge down.** The forge rows become `?` and the
  cursor stays where it was.
- **A check renamed upstream.** Its table row stops matching, and the
  fallback prints the check's sentence — the test below names it.

## Tests required

Unit, `internal/screen`: the selection skips headings and blanks; the
footer's text for a met, an unmet and an unnamed requirement.

Unit, `internal/command/doctorcmd`: every requirement name the binary
can report has a row in the embedded table. This is the test that keeps
the document and the checks in step.

Integration: the rendered frame matches the drawing's rows and footer.

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`adoption/doctor.excalidraw`](../../docs/product/screens/adoption/doctor.excalidraw)
  — `the doctor screen — a requirement that is met` and `the doctor
  screen — a requirement that is not`.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [ ] Every requirement row is selectable and explained.
- [ ] The table is written, embedded, and covers every name.
- [ ] A requirement with no row falls back to the check's sentence.

## Proposed product changes

- `product/adoption/doctor.md`
  — the table of requirements and their explanations.

## Proposed technical changes

- `technical/engineering/coupling.md`
  — a fifth rule: a sentence a product doc states is embedded from that
  doc, never retyped into Go beside it.

## Outcome

_(fill after execution)_

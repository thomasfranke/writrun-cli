---
id: spec-0037
task_ref: task-0033
status: implemented
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

- `product/adoption/doctor.md` — the table of requirements and their
  explanations.

## Proposed technical changes

- `technical/engineering/coupling.md` — a fifth rule: a sentence a
  product doc states is embedded from that doc, never retyped into Go
  beside it.

## Outcome

`internal/screen/requirements.go` is the doctor screen: every
requirement row selectable, the met ones included, with the headings,
the counts and the blank lines shown and skipped. The footer is a rule,
`requirement — what it is` under a hanging indent, then the keys. `r`
re-runs every check and keeps the cursor. `writrun doctor` opens it
where stdin and stdout are terminals and prints the report otherwise,
the split `config` already makes.

`adoption/doctor.md` gained the table: one row per requirement — what
it is, why the stage needs it, what clears it — and
`internal/command/doctorcmd/explanations.go` holds that table and reads
it. A requirement the table does not name falls back to the check's own
sentence and nothing else. Both screen frames are asserted against what
the model renders; the rows are compared line for line and the footer
as one paragraph, because where it wraps is the terminal's width.

**What the plan did not foresee.**

- **"Embed that table at build time" is not available here.**
  `//go:embed` reads only files at or under its own package directory,
  and the document is three directories above every package. Embedding
  it would take a Go package at the repository root, which
  `technical/layout/public-surface.md` forbids and this spec promised no
  change to. The table is held in `explanations.go` line for line
  instead, and `TestTheTableInGoIsTheDocumentsOwn` fails unless the
  document still carries it verbatim — so the two cannot disagree in
  silence, which is what the step asked for. `coupling.md`'s new rule
  states the constraint and the substitute.
- **Two tests keep the document and the checks in step, not one.** Every
  name the binary reports has a row, and every row names a requirement
  some check makes — the second catches a row left behind by a renamed
  check, which the first cannot see.
- **The drawing disagrees with itself once.** Its footer quotes the
  selected row as `8 rows, 8 answered` where the row it points at is
  drawn `8 gates, 8 answered`. The footer quotes the row's note, so the
  row is the authority; the case says so where it reconciles the word.
- **The coupling guard fired on the copied table.** `writrun/gates.md`
  and `writrun/settings.json` appear in it as the document's own words.
  `internal/kit/coupling_test.go` now passes over a string literal that
  is a markdown table row: nothing dereferences it, and a tag that moved
  the file turns the lookup's own test red.
- **`doctor` no longer declares `AsksNothing`**, because it opens a
  terminal program. From the entry screen it is spawned as a process of
  its own, which is what decision 0015 already prescribes for a command
  that takes the keyboard.

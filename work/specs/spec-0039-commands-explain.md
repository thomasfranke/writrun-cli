---
id: spec-0039
task_ref: task-0035
status: implemented
created: 2026-09-13T22:08:07Z
---

# spec-0039 — A command run bare explains itself

**References:** [task-0035](../tasks/task-0035-binary-explains.md)

- **Goal:** A reader who has not read the methodology can learn what a command is for from the command itself.

## Scope

In: the long description each command carries, when it is printed, the
plainer one-liners, and the grouping of `--help`.

Out: what any command does; the footers and cursors
([spec-0040](spec-0040-plan-navigation.md)).

## Steps

1. Give the command table a long description beside the one-liner it
   already holds: one plain sentence, then *why you would*, *what it
   does*, *what it leaves* or *what it never*, and *what comes next*.
2. Print it from `writrun <command> --help`, for every command.
3. Print it also on a bare interactive run of a command that would
   otherwise ask or refuse — `take`, `init`, `amend`, `author`,
   `finish`, `report`, `update`, `uninstall`, `config` — before its
   first question.
4. Never print it on a bare run of a command that does its work without
   an argument: `list`, `status`, `doctor` and `work` are run daily, and
   a daily explanation is noise.
5. Rewrite the one-liners in plain words, and read them from the table
   on both surfaces that show them — the entry screen and `--help`.
6. Group `--help`'s rows by what a person is doing: getting started,
   doing the work, writing the rules, keeping it current.

## Acceptance criteria (EARS)

- When a command is run with `--help`, the system shall print its long
  description.
- When a command that asks is run bare on a terminal, the system shall
  print its long description before its first question.
- When `list`, `status`, `doctor` or `work` is run bare, the system
  shall do its work and print no description.
- When any argument is given, the system shall print no description.
- When `--help` is run, the rows shall be grouped and their text shall
  be the table's own.
- When a command is added without a long description, the build shall
  fail.

## Edge cases

- **No terminal.** The description is skipped: a script reading a
  command's output keeps reading what it read before.
- **`--yes` on a command that asks.** The description prints, the
  question does not.
- **A command reached from the entry screen.** The row's one-liner and
  `--help`'s are the same string, so they cannot drift.

## Tests required

Unit, `internal/command`: every command has a long description (a table
test over the registry); `--help` prints it; a bare run prints it only
for the commands that ask.

Integration, `tests/integration/cli/`: `writrun --help` is grouped;
`writrun amend` explains before asking; `writrun list` does not.

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`help.excalidraw`](../../docs/product/screens/help.excalidraw)
  — `writrun --help — grouped by what a person is doing`.
- [`screens/`](../../docs/product/screens/README.md)
  — the frame `writrun <command> — what it is for, in its own words`,
  one in each of the thirteen commands' own drawings.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [x] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [x] Every command carries a long description, checked by a test.
- [x] The one-liners are plain, and read from one place.
- [x] `--help` is grouped.

## Proposed product changes

- `product/rules.md` — how a command reports, which today says
  `--help` restates nothing.

## Proposed technical changes

- none — no machinery change.

## Outcome

`command.Command` carries an `About`: a sentence and the labelled parts
under it, each as the lines the drawing draws. `writrun <command>
--help` prints it for all thirteen, and `dispatch` prints it before the
work of a command run with no arguments where stdin is a terminal —
`AsksNothing` is already the set of four that print none, so no second
list says which commands explain themselves. `--help` is four groups,
named in `internal/command/run.go`, and every row's text is the
command table's summary the entry screen also reads.

`internal/drawing` reads a frame out of a drawing at test time, so the
assertion is the drawing and never a transcription of it.
`cmd/writrun` compares the production table's `--help` with the
`help.excalidraw` frame and each command's with its own, line for line.

Six things the plan did not foresee.

The long description is stored as the lines the drawing draws, not as
paragraphs a wrapper breaks. The drawn wraps are a hand's, at no width
the code could compute, and the frames are the assertion.

`--help`'s last two lines are drawn below the frame's divider, where
that drawing puts what the annotation beside it says — so the frame
assertion covers the header and the rows, and the two lines under them
are held by a case in `internal/command`.

The terminal that decides a bare run's description is stdin's, not
stdout's: the description precedes a question, a question is asked on
stdin, and `WRITRUN_TTY_IN` is the seam the suite already stands a
terminal up with.

"When a command is added without a long description, the build shall
fail" is answered by a test over the production table, not by the
compiler. Go cannot require a struct field, and a constructor that
could would have rewritten all thirteen command packages to enforce
what one table test states.

The entry screen's drawing still states the one-liners this change
replaced, and bounds a summary at 44 characters where its own window
gives 66 — [report-0042](../reports/report-0042-stale-drawn-summaries.md).

[decision 0010](../../docs/technical/decisions/docs/0010-help-is-one-line-per-command.md)
records `--help` as one line per command and names full help text among
what it rejected — [report-0043](../reports/report-0043-decision-rejects-long-help.md).
This spec promises `product/rules.md` and nothing under
`technical/decisions/`, so the decision is a report and not an edit
made here.

The Proposed-changes entry was rewritten from a markdown link to the
bare backticked path that both gates read
([report-0041](../reports/report-0041-invisible-promises.md)). In the
form it carried, `check_deltas.sh` read this spec as promising nothing
and would have refused `product/rules.md` as undeclared.

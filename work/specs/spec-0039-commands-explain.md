---
id: spec-0039
task_ref: task-0035
status: draft
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

## Definition of Done

- [ ] Every command carries a long description, checked by a test.
- [ ] The one-liners are plain, and read from one place.
- [ ] `--help` is grouped.
- [ ] `help.excalidraw` and every command drawing's explaining frame
      re-transcribed.

## Proposed product changes

- [`product/screens/help.excalidraw`](../../docs/product/screens/help.excalidraw)
  — both frames, re-transcribed.
- Every drawing under
  [`product/screens/`](../../docs/product/screens/README.md) — the
  explaining frame each one carries.
- [`product/rules.md`](../../docs/product/rules.md) — how a command
  reports, which today says `--help` restates nothing.

## Proposed technical changes

- `internal/command` — the long description on the command type, and
  the test that requires it.

## Outcome

_(fill after execution)_

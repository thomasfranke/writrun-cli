---
id: spec-0040
task_ref: task-0034
status: draft
created: 2026-09-13T22:08:09Z
---

# spec-0040 — A cursor and a footer on every plan and every question

**References:** [task-0034](../tasks/task-0034-screen-furniture.md)

- **Goal:** Every plan and every question puts a cursor on its rows and names the one under it, with the act the next key performs.

## Scope

In: the description under a navigated question, the cursor and footer
on the plans `take`, `finish`, `author`, `amend`, `update` and
`uninstall` show, and the queue screen's own footer.

Out: the explanations themselves ([spec-0039](spec-0039-commands-explain.md));
the doctor screen ([spec-0037](spec-0037-doctor-screen.md)).

## Steps

1. Give `internal/term`'s question a description per option, rendered
   under the list. Every navigated question inherits it at once — the
   task in `take`, the stage in `init`, the key in `config`.
2. Render the plans through one component: the rows the command already
   composes, a cursor, and a footer naming the selected row.
3. Say in the footer what the next key does to that row, not only what
   the row is.
4. End a plan's footer `esc cancels`, a screen's `esc back · q quit` —
   the furniture the entry screen defines.
5. Leave the no-terminal path exactly as it is: the plan prints, the
   flags answer, and nothing navigates.

## Acceptance criteria (EARS)

- When a question is navigated, the system shall show the highlighted
  option's description.
- When a plan is shown on a terminal, its rows shall be selectable and
  the footer shall name the selected one.
- When a row is selected, the footer shall name the act the next key
  performs on it.
- When there is no terminal, the output shall be what it is today.
- When `--yes` is given, the system shall navigate nothing.

## Edge cases

- **A plan with one row.** The cursor rests on it; the footer explains
  it.
- **A plan longer than the terminal.** It scrolls, and the footer stays.
- **A question with no description for an option.** It shows the option
  alone rather than an empty pane.

## Tests required

Unit, `internal/term`: the description renders under the highlighted
option.

Unit, `internal/screen`: the plan component's rows, cursor and footer.

Integration: `take` with a fake terminal shows the description; without
one, its output is byte-for-byte what it is today.

## Definition of Done

- [ ] One question shape, one plan component — no second copy.
- [ ] The no-terminal output is unchanged, proved by a test.
- [ ] Every drawing's navigating frame re-transcribed.

## Proposed product changes

- The navigating frames in
  [`tasks/`](../../docs/product/screens/tasks/),
  [`authoring/`](../../docs/product/screens/authoring/) and
  [`adoption/`](../../docs/product/screens/adoption/) — re-transcribed.

## Proposed technical changes

- `internal/term` — the per-option description.
- `internal/screen` — the plan component the four pull-request commands
  and the two adoption commands share.

## Outcome

_(fill after execution)_

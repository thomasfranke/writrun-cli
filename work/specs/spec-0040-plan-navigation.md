---
id: spec-0040
task_ref: task-0034
status: implemented
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

The drawing is the assertion. These are the frames, each found in its
file by the caption drawn above it:

- [`tasks/take.excalidraw`](../../docs/product/screens/tasks/take.excalidraw)
  — `writrun take — the available group, a task highlighted`.
- [`adoption/init.excalidraw`](../../docs/product/screens/adoption/init.excalidraw)
  — `writrun init — the stage question, a stage highlighted`.
- [`tasks/finish.excalidraw`](../../docs/product/screens/tasks/finish.excalidraw)
  — `writrun finish — the summary, one line selected`.
- [`authoring/author.excalidraw`](../../docs/product/screens/authoring/author.excalidraw)
  — `writrun author — the composition, one line selected`.
- [`authoring/amend.excalidraw`](../../docs/product/screens/authoring/amend.excalidraw)
  — `writrun amend — the amendment, one line selected`.
- [`adoption/update.excalidraw`](../../docs/product/screens/adoption/update.excalidraw)
  — `writrun update — the plan, one path selected`.
- [`adoption/uninstall.excalidraw`](../../docs/product/screens/adoption/uninstall.excalidraw)
  — `writrun uninstall — the plan, one line selected`.
- [`tasks/list.excalidraw`](../../docs/product/screens/tasks/list.excalidraw)
  — `the queue screen — a task selected`.

Each is compared with what the binary renders, and a disagreement
fails the suite. The drawing is not the thing under test: it is what
the test is written against.

## Definition of Done

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
      Six of the eight hold whole. `tasks/take.excalidraw` and
      `tasks/list.excalidraw` are held by their rows and their keys:
      their panes name a task's spec, which the lister's row does not
      carry ([report-0045](../reports/report-0045-panes-unread-facts.md)).
- [x] One question shape, one plan component — no second copy.
- [x] The no-terminal output is unchanged, proved by a test.

## Proposed product changes

- none — the drawings already state this, and the code comes to them.
  A drawing that must itself change is a documentation change of its
  own, decided by a person before any code answers it.

## Proposed technical changes

- none — no machinery change.

## Outcome

Built, as one component. `internal/screen/choose.go` renders a list of
rows with a cursor, the selected row's own sentence under them, and a
footer ending `esc cancels`. A question and a plan are the same screen
because they are the same act, so there is one of them and not two:
`internal/term` turns a `command.Option` and a `command.Plan` into it,
and the two questions — the stage in `init`, the task in `take` — went
with it off `huh`, whose field renders its one description above the
options where every drawing puts the highlighted option's own sentence
below them.

Every plan goes through `Ctx.AskPlan`. Five commands compose rows
where they printed lines — `finish`, `author`, `amend`, `update`,
`uninstall` — and `take` hands over what `take_task.sh` composed,
line for line, because which line means what is the script's answer
and not this binary's. The no-terminal path prints those same rows and
asks the same yes/no, so its bytes are what they were.

All eight frames are asserted, each beside the command that composes
it (`internal/command/*/frames_test.go`), through `drawing.Screen` and
`drawing.Compare` — one reader and one comparison for seven packages.

**What the plan did not foresee.**

`Plan` carries two forms, not one. `author`'s screen counts the body
where the printed form quotes it, and `uninstall`'s rows name a path
where the printed rows carry the clause about it — both because the
drawings state the short form and step 5 states the printed one. A
`Printed` field beside `Rows` is what keeps those from being two
compositions; where a command needs only one form it gives only `Rows`.

Two frames explain a row with facts the row does not carry.
`tasks/take.excalidraw` says a task's spec is approved and its priority
is medium, and `tasks/list.excalidraw` adds `no spec`; the lister's own
line — `printf '  %-10s %-7s %s'` — carries the id, the priority and
the title, and nothing else. Those two are asserted by their rows and
their keys, and the difference is
[report-0045](../reports/report-0045-panes-unread-facts.md).

`authoring/author.excalidraw` disagrees with itself by one column: its
selected row sits one column right of the three rows above it. The
three are the authority and the case reconciles it, as the doctor
screen's case already reconciles a word.

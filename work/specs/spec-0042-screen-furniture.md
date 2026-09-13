---
id: spec-0042
task_ref: task-0034
status: draft
created: 2026-09-13T22:29:36Z
---

# spec-0042 — One header, one footer, and one way out on every screen

**References:** [task-0034](../tasks/task-0034-screen-furniture.md)

- **Goal:** Every screen opens with the same two lines, ends with the same one, and gives every key the same meaning it has on the screen before it.

## Scope

In: the header, the footer, the pager's own furniture, and the sentence
an unreadable read prints.

Out: what any screen lists — the rows are each screen's own.

## Steps

1. Give every screen a two-line header: the identity line
   `cmd/writrun/main.go`'s `header()` already composes, and a context
   line naming what this screen reads. The queue screen has neither
   today.
2. End every screen with one footer in three groups, in this order:
   movement, the actions one key each, the way out.
3. `esc` goes back and `q` quits. A screen with nowhere to go back to —
   the entry screen, the first run — offers `q quit` alone. The config
   screen's `q back`, which quits nothing, goes
   ([report-0035](../reports/report-0035-screen-furniture.md)).
4. Wrap a dispatched command's output in the same furniture: the pager
   is a screen. Run from a shell the command prints its own lines and no
   header, as [rules.md](../../docs/product/rules.md) requires of plain
   output.
5. Leave a question without header or `q` — `q` would be a value — and
   ending in `internal/term`'s `wayOut`, unchanged.
6. Make a failed read name the command that answers it, not only the
   error.

## Acceptance criteria (EARS)

- When any screen renders, it shall open with the identity line and a
  context line.
- When a screen has somewhere to go back to, its footer shall end
  `esc back · q quit`.
- When it has not, its footer shall end `q quit`, and shall not offer
  `esc`.
- When a command's output is shown in the pager, the pager shall carry
  the same header and the same way out.
- When a question is asked, it shall carry no header and no `q`.
- When a read fails, the message shall name the command that answers it.
- When a command runs from a shell, its output shall carry no header.

## Edge cases

- **A terminal narrower than the header.** The header truncates; it
  never wraps into the rows.
- **The settings screen with no settings file.** Its own state, drawn:
  it says so and offers no edit.
- **A screen with nothing selectable.** The footer drops the keys that
  cannot act and says `nothing to select`, as the queue already does.

## Tests required

Unit, `internal/screen`: each screen's first two lines and its last;
a table test asserting every footer ends in one of the two forms and
that no screen composes its own header.

Integration, `tests/integration/screen/`: the rendered frames match the
drawings, including the pager's.

The drawing is the assertion. Each frame named above is read out of its
`.excalidraw` and compared with what the binary renders, and a
disagreement fails the suite. The drawing is not the thing under test:
it is what the test is written against.

## Definition of Done

- [ ] The frames named above are asserted against what the binary
      renders. Where the two disagree, the binary is what changes.
- [ ] Header and footer come from one place; no screen writes its own.
- [ ] `q` quits on every screen, and `esc` is offered only where there
      is somewhere to go.
- [ ] [report-0035](../reports/report-0035-screen-furniture.md) closed.
- [ ] `entry`, `tasks/list`, `adoption/config` and
      `adoption/doctor` re-transcribed.

## Proposed product changes

- [`product/screens/README.md`](../../docs/product/screens/README.md) —
  the furniture every screen carries, stated as a rule.

## Proposed technical changes

- none — no machinery change.

## Outcome

_(fill after execution)_

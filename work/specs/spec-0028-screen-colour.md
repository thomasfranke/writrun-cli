---
id: spec-0028
task_ref: task-0029
status: draft
created: 2026-09-08T15:04:00Z
---

# spec-0028 — Colour the screens the reporting rule already promises

**References:** [task-0029](../tasks/task-0029-ui-screens-refactor.md)

- **Goal:** The colour rule in
  [`rules.md`](../../docs/product/rules.md) becomes visible output.

## Scope

In: a palette, and its use on the screens and in the reports the
drawings colour — the entry screen's groups and stage block, the queue's
sections, `doctor`'s finding levels.

Out: colouring output no drawing colours. A command whose drawing is
monochrome stays monochrome.

## Steps

1. Give `Ctx.Color` a consumer. It is computed in
   `internal/command/run.go` and read nowhere, so the rule it implements
   has never reached a pixel.
2. Define the palette once, beside the terminal port, in the terms the
   drawings use: a group heading, a declared setting, a finding that
   breaks, one that advises, one unread, and dimmed text.
3. Apply it where the drawings apply it, and nowhere else.
4. Carry the decision through `doctor`'s three levels, which already
   exist in `internal/command/doctorcmd/report.go` and print as words in
   one column.

## Acceptance criteria (EARS)

- When stdout is a terminal and neither `NO_COLOR` nor `--no-color` is
  given, the system shall colour the elements the drawings colour.
- When stdout is not a terminal, the system shall emit no escape
  sequence.
- When `NO_COLOR` is set, the system shall emit no escape sequence.
- When `--no-color` is given, the system shall emit no escape sequence.
- When colour is disabled by any of the three, the system shall print
  the same words in the same columns as when it is enabled, so no
  meaning rides on colour alone.
- When `doctor` prints a finding, the system shall colour it by its
  level and shall keep the level's word in its column.

## Edge cases

- **A terminal that reports colour support it does not have.** The rule
  is the three conditions already in `colorEnabled`, and no capability
  probe is added: a fourth condition would be a fourth answer to a
  question `rules.md` has settled.
- **Output redirected mid-pipeline.** `Ctx.Color` is decided once, at
  the start, from the same probe every command shares.
- **A screen rendered by Bubble Tea rather than printed.** The screen
  owns the terminal while open, so it takes the same decision from the
  same field rather than probing again.

## Tests required

- Unit, `internal/command`: `colorEnabled` over the matrix of terminal,
  `NO_COLOR` and `--no-color` — extending what exists rather than
  restating it.
- Unit: with colour off, the rendered text of each screen is byte-equal
  to today's.
- Unit: with colour on, no line loses a word or shifts a column against
  the colour-off rendering.
- Integration: a piped run of `doctor` and of `list` contains no `\x1b`.

## Definition of Done

- [ ] `Ctx.Color` has a consumer.
- [ ] The palette is defined in one place.
- [ ] Every element the drawings colour is coloured; nothing else is.
- [ ] No meaning is carried by colour alone.

## Proposed product changes

- `product/rules.md` — the colour rule gains the sentence that
  colour never carries meaning alone, which is what makes the other
  three conditions safe.

## Proposed technical changes

- `technical/layout/public-surface.md` — the palette's home, named
  where the terminal port is named.

## Outcome

_(fill after execution)_

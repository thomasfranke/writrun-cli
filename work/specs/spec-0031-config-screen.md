---
id: spec-0031
task_ref: task-0029
status: draft
created: 2026-09-08T15:04:03Z
---

# spec-0031 — A config screen edits settings through the kit's checker

**References:** [task-0029](../tasks/task-0029-ui-screens-refactor.md)

- **Goal:** `.writrun/settings.json` is editable from the binary without
  the binary holding a second copy of its schema.

## Scope

In: showing the settings, changing one, and the rule that permits it.

Out: knowing which values are legal. The binary writes and asks the kit;
it never decides. Also out: the stage's consequences — declaring a stage
does not install anything, and `update` is still what fetches the kit.

## Steps

1. Read every value through the kit's own
   `.writrun/scripts/stage-2-pull-requests/read_setting.sh`.
2. Show the keys grouped by the section that owns them, as
   [`config.excalidraw`](../../docs/product/screens/config.excalidraw)
   draws them.
3. On a change: write the new value, then run the kit's own
   `.writrun/scripts/stage-2-pull-requests/check_settings.sh`. Keep the
   write only on exit 0.
4. On a non-zero exit, restore the file byte for byte and print the
   script's own refusal. That refusal names the vocabulary, which is how
   a reader learns the legal values without this binary storing them.
5. Amend [`rules.md`](../../docs/product/rules.md) so the rule against
   overwriting the project's own files says what it has always meant:
   the kit does not overwrite the adopter, and the adopter may change
   their own settings through the porcelain.

## Acceptance criteria (EARS)

- When the config screen opens, the system shall show every key the kit
  documents, under the section the kit gives it.
- When the config screen opens, the system shall read values through the
  kit's reader and shall hold no list of keys of its own.
- When a value is changed and the kit's checker exits 0, the system
  shall keep the change.
- When a value is changed and the kit's checker exits non-zero, the
  system shall restore the file to its previous bytes and print the
  checker's own message.
- When the file is restored after a refusal, the system shall leave no
  other change on disk.
- When a change would be written, the system shall show the key, the old
  value and the new one, and ask, as every changing command does.
- When `--yes` is given, the system shall skip the question and not the
  check.
- When the settings file is absent, the system shall say so and offer no
  edit: the kit's reader documents defaults, and writing a file the
  project never had is adoption's act, not this one.

## Edge cases

- **A key the kit documents that this repository's file omits.** The
  reader returns the documented default; the screen shows it as the
  default and an edit writes it in the section the kit names.
- **A key in the file that the kit does not document.** The checker
  already refuses it. The screen shows what the checker says rather than
  hiding the key.
- **The checker absent** — a stage-1 adoption may not carry it. Without
  the checker there is nothing to validate against, so the screen shows
  values and refuses edits, naming why.
- **Two writes racing.** The write, the check and the restore are one
  motion per change; a second change is not begun until the first has
  settled.
- **A change to `stage`.** It is a declaration, not an installation. The
  screen says so, because a reader who raises the stage will otherwise
  expect scripts to appear.

## Tests required

- Unit: a change accepted by a stubbed checker is kept; one refused is
  restored byte for byte and the checker's message reaches the user.
- Unit: the key list comes from the kit, not from a Go literal — a
  fixture kit with an extra key shows that key.
- Unit: `--yes` skips the question and still runs the checker.
- Integration, a new `tests/integration/config/`: a real settings file
  edited to a legal value survives the real checker; an illegal one
  leaves the file unchanged and exits non-zero.

## Definition of Done

- [ ] No allowed value is written in Go.
- [ ] A refused edit leaves the file exactly as it was.
- [ ] `rules.md` permits the edit in words.
- [ ] The screen matches the drawing.

## Proposed product changes

- `docs/product/rules.md` — the overwrite rule gains the adopter's own
  settings as its stated exception.
- `docs/product/config.md` — the command's page: what it shows, what it
  writes, and what it never decides.
- `docs/product/README.md` — its row in the flow table.
- `docs/product/screens/README.md` — the table drops `Proposed` from
  Config.
- `docs/product/screens/config.excalidraw` — the `proposed` wording
  goes.

## Proposed technical changes

- `docs/technical/engineering/coupling.md` — the settings schema joins
  the list of things the kit owns and this binary reads rather than
  restates.

## Outcome

_(fill after execution)_

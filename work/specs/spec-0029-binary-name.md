---
id: spec-0029
task_ref: task-0029
status: implemented
created: 2026-09-08T15:04:01Z
---

# spec-0029 — The binary names itself writrun-cli

**References:** [task-0029](../tasks/task-0029-ui-screens-refactor.md)

- **Goal:** One name in everything the binary says about itself, and a
  recorded decision for why it is no longer the one
  [0005](../../docs/technical/decisions/runtime/0005-the-binary-is-writrun.md)
  chose.

## Scope

In: what the binary calls itself in `--version`, `--help` and the
screens, and the decision that supersedes 0005.

Out: the executable's filename and the command a person types, which
stay `writrun`. The repository, the module and the Homebrew formula are
already `writrun-cli` and are untouched.

## Steps

1. Write the decision under
   `docs/technical/decisions/runtime/`, stating that the product is
   named `writrun-cli` and the command stays `writrun`, and that it
   supersedes 0005 — which rejected a second name when the second name
   was `writ`, a shortening that bought nothing. The case here is the
   opposite: `writrun` and `WritRun` differ by capitalisation alone, and
   a reader cannot hear the difference.
2. Amend 0005 to name its successor, as
   [0012](../../docs/technical/decisions/docs/0012-technical-subjects-are-folders.md)
   does for 0006.
3. Change `--version` to print the product name.
4. Change `--help`'s first line to the same name.
5. Leave every invocation, path and import as it is.

## Acceptance criteria (EARS)

- When `--version` runs, the system shall print `writrun-cli`, the
  client's version, and the WritRun tag it pins.
- When `--help` runs, the system shall name the product `writrun-cli`.
- When a screen prints its header, the system shall use the same name
  `--version` prints.
- When a command is invoked, the system shall answer to `writrun` as it
  does today.
- When the decision is read, the system shall state which decision it
  supersedes, and 0005 shall name it back.

## Edge cases

- **A user searching for `writrun` in the help.** The command names in
  the help body are unchanged; only the product line changes.
- **A source build with no tag.** The version is the Go pseudo-version,
  38 characters. `--version` prints it whole, as today; the screen
  header shortens it — that is
  [spec-0027](spec-0027-entry-screen.md)'s header, not this spec's
  string.
- **The Homebrew formula and the release assets.** Both are already
  named for the repository and need no change; a rename there would be a
  distribution change, not this one.

## Tests required

- Unit, `internal/command`: `--version` and `--help` name the product.
- Integration, `tests/integration/cli/`: the compiled binary answers to
  `writrun` and names itself `writrun-cli`.
- The release suite is unchanged and must stay green, which is what
  proves the rename did not reach distribution.

## Definition of Done

- [ ] One name in `--version`, `--help` and the screens.
- [ ] The command a person types is unchanged.
- [ ] A decision file records the change and supersedes 0005.
- [ ] 0005 names its successor.

## Proposed product changes

- `product/rules.md` — the `--version` rule names the product it
  prints.

## Proposed technical changes

- `technical/decisions/runtime/0005-the-binary-is-writrun.md` —
  names the decision that supersedes it.
- `technical/decisions/runtime/` — the new decision file.
- `technical/decisions/README.md` — its index row.

## Outcome

Built as planned. `--version` and `--help` name `writrun-cli`, the
command a person types is unchanged, and
[0014](../../docs/technical/decisions/runtime/0014-the-product-is-writrun-cli.md)
records why 0005 cedes — it rejected a second name when that name was
`writ`, and the case here runs the other way: `writrun` and `WritRun`
differ by capitalisation alone, so a screen naming both reads as one
thing twice. 0005 names its successor, as 0012 does for 0006.

One thing the plan did not say: the name needed a home. It is
`command.Product`, exported, because the screen's header needed it too
and a second copy of a name is a second name.

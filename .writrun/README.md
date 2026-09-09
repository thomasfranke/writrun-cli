# .writrun — WritRun's home in this repository

Everything in this folder is WritRun's, whole: `writ update` replaces
it entire, and nothing here is the adopter's to edit. The project's own
answers live in `writrun/` — `settings.json`, `gates.md`,
`conventions/` — which no update touches. The split is the two homes
rule: every file is exactly one side's, so an update never merges
anyone's prose.

What the platform dictates stays where the platform demands — the five
`writrun-*.yml` workflows in `.github/workflows/` (GitHub runs them
nowhere else), the four-line pointer WritRun grafts into the adopting
project's root `AGENTS.md` (agents look for it there, and it is all
WritRun claims of that file), and `WRITRUN.md` at the root, the guide
written for humans — each declaring in its own header that WritRun
shipped it.

| | Holds | `writ update` |
|---|---|---|
| `AGENTS.md` | the agent flow the root pointer names | replaces whole |
| `skills/` | the `writrun-*` skills and their scripts | refreshes |
| `scripts/` | the step logic, testable bash, one folder per adoption stage: from Stage 2 the workflows call it (`gh` wherever the forge must be asked), and the Stage 1 folder holds what a person or an agent runs directly | refreshes |
| `templates/` | shipped default body shapes for task, spec and report, and the PR body template (its only home: agents fill it when opening PRs; GitHub's pre-fill is deliberately forgone) | refreshes |
| `VERSION` | the tag this copy of the kit came from | rewrites |

Two rules keep the layers honest:

- **The project's file always wins.** A body shape in
  `writrun/conventions/templates/` beats `templates/`; a convention the
  project rewrote is the convention. Nothing in this folder is
  authority over the project's own.
- **Never hand-edit this folder** — customize in `writrun/` instead.
  Hand edits here are overwritten by the next refresh, by design. A kit
  change worth having is an issue on the WritRun repository, per the
  flow in [`AGENTS.md`](AGENTS.md).

Adopting a project? The kit ships as `kit/` in WritRun's repository,
shaped exactly like the destination root, and its guide travels with the
copy as `WRITRUN.md`. In WritRun's own repository, `kit/` is held
byte-identical to the root by a unit test; maintainers refresh it with
`make kit-sync`.

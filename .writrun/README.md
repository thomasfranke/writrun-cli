# .writrun — WritRun's home in this repository

Everything WritRun ships that is WritRun's own to place lives here, so
provenance is never in doubt. What the platform dictates stays where the
platform demands — the four `writrun-*.yml` workflows in
`.github/workflows/` (GitHub runs them nowhere else), `AGENTS.md` at the
root (agents look for it there) — and each of those declares in its own
header that WritRun shipped it.

| | Owned by | `writ update` |
|---|---|---|
| `skills/` | WritRun — the `writrun-*` skills and their scripts | refreshes |
| `scripts/` | WritRun — the step logic, testable bash, one folder per adoption stage: from Stage 2 the workflows call it (`gh` wherever the forge must be asked), and the Stage 1 folder holds what a person or an agent runs directly | refreshes |
| `templates/` | WritRun — shipped default body shapes for task, spec and report, and the PR body template (its only home: agents fill it when opening PRs; GitHub's pre-fill is deliberately forgone) | refreshes |
| `VERSION` | WritRun — the tag this copy of the kit came from | rewrites |
| `conventions/` | **The project**, from the moment of adoption — ships as defaults, then it is yours | never touches |
| `settings.json` | **The project**, from the moment of adoption — the stage, the conduct flags, the title style; the first file to edit after adoption | never touches |

Two rules keep the layers honest:

- **The project's file always wins.** A body shape in
  `conventions/templates/` beats `templates/`; a convention you rewrote is
  the convention. Nothing in a WritRun-owned folder is authority.
- **Never hand-edit a WritRun-owned folder** — customize at the project
  layer instead. Hand edits there are overwritten by the next refresh, by
  design. The two project-owned rows above are the exception, and
  `settings.json` is the one *file* among them: it lives at this folder's
  root rather than in `conventions/` because the one address ends the
  hunt, and editing it is the point.

Adopting a project? The kit ships as `template/` in WritRun's repository,
shaped exactly like the destination root, and its guide travels with the
copy as `WRITRUN.md`. In WritRun's own repository, `template/` is held
byte-identical to the root by a unit test; maintainers refresh it with
`make template-sync`.

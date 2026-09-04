# `writrun init`

Installs the WritRun kit into an existing repository.

- Runs only where `.writrun/` is absent; refuses an already-adopted
  repository and points at [`update`](update.md) instead.
- Installs the kit at a pinned WritRun tag and records which tag it
  took.
- **Extracts the repository's existing conventions** — its commit
  history, its contributing guide — into the conventions folder, rather
  than imposing the shipped defaults.
- **Grafts an existing `AGENTS.md`**, never overwrites it: WritRun's
  part enters as one fenced section. A repository without one gets the
  skeleton.
- Installs a commit-message hook that **validates** the commit
  convention. It never writes a message — the message belongs to
  whoever made the change.
- **Asks the stage** — 1 (files only), 2 (pull requests), 3 (GitHub
  issues) — writes the answer to `.writrun/settings.json`, and runs the
  chosen stage's [`doctor`](doctor.md) checks on the spot. What is
  missing is named, never fixed: adoption is not conditioned on the
  forge.
- Leaves the queue empty. Work arrives through the pipeline, never from
  installation.

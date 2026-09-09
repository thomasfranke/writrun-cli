# `writrun init`

Installs the WritRun kit into an existing repository.

- Runs only where `.writrun/` is absent; refuses an already-adopted
  repository and points at [`update`](update.md) instead.
- Installs the kit at a pinned WritRun tag and records which tag it
  took.
- **Extracts the repository's existing conventions** — its commit
  history, its contributing guide — and writes the vocabulary it found
  where the checks read it, rather than imposing the shipped defaults.
  A convention explains a vocabulary; it never carries a second copy
  for a check to disagree with.
- **Grafts an existing `AGENTS.md`**, never overwrites it: WritRun's
  part enters as one section, the heading that links
  `.writrun/AGENTS.md`. A repository without one gets the skeleton, and
  one already carrying the link is left alone.
- Installs a commit-message hook that **validates** the commit
  convention. It never writes a message — the message belongs to
  whoever made the change.
- **Asks the stage.** The three are named here, and named nowhere else:
  **1 files**, **2 pull requests**, **3 GitHub issues**. It writes the
  answer to the settings and runs the chosen stage's
  [`doctor`](doctor.md) checks on the spot. What is
  missing is named, never fixed: adoption is not conditioned on the
  forge.
- Leaves the queue empty. Work arrives through the pipeline, never from
  installation.

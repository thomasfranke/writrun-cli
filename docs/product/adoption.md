# Adoption — `init`, `update`, `doctor`, `uninstall`

## `writrun init`

Installs the WritRun kit into an existing repository.

- Runs only where `.writrun/` is absent; refuses an already-adopted
  repository and points at `update` instead.
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
- Leaves the queue empty. Work arrives through the pipeline, never from
  installation.

## `writrun update`

Refreshes an adopted kit to a newer WritRun tag.

- Refreshes only what the methodology declares refreshable: the kit's
  own files, and the fenced section of `AGENTS.md`.
- **Preserves the lines the methodology marks as the project's** —
  the human-gates table among them — across every refresh.
- Never touches the conventions folder, the project's docs, or the
  queue.
- Shows what will change before changing it, and stops if the fenced
  markers are missing or damaged.

## `writrun doctor`

Reports whether the repository still satisfies what the methodology
assumes.

- Checks that the fenced markers survived editing, that the kit's
  version is recorded, and that the queue is readable.
- Checks that the forge settings the flows depend on match what the
  project declares.
- Reports; it never repairs. Every finding names the file or setting
  and what is expected of it.
- Exit status is non-zero when any finding would break a flow.

## `writrun uninstall`

Removes the WritRun kit from an adopted repository.

- Runs only where `.writrun/` is present; refuses a repository that was
  never adopted.
- Removes what `init` installed: the kit's files, the commit-message
  hook, and the fenced WritRun section of `AGENTS.md` — the rest of
  that file stays exactly as the project wrote it.
- **Never touches `work/`.** Tasks, specs, and reports are the
  project's record, not the kit's — the queue and its history survive
  the tooling that managed them.
- Never touches the project's own docs. What the methodology helped
  write belongs to the repository, not to the kit.
- Shows what will be removed and what will stay, and asks for
  confirmation before removing anything.

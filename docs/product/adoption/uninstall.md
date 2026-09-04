# `writrun uninstall`

Removes the WritRun kit from an adopted repository.

- Runs only where `.writrun/` is present; refuses a repository that was
  never adopted.
- Removes what [`init`](init.md) installed: the kit's files, the
  commit-message hook, and the fenced WritRun section of `AGENTS.md` —
  the rest of that file stays exactly as the project wrote it.
- **Never touches `work/`.** Tasks, specs, and reports are the
  project's record, not the kit's — the queue and its history survive
  the tooling that managed them.
- Never touches the project's own docs. What the methodology helped
  write belongs to the repository, not to the kit.
- Shows what will be removed and what will stay, then asks
  ([rules](../rules.md)).

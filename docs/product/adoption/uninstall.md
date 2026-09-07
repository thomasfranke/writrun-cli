# `writrun uninstall`

Removes the WritRun kit from an adopted repository.

- Runs only where `.writrun/` is present; refuses a repository that was
  never adopted.
- Removes what [`init`](init.md) installed: the kit's files, the
  commit-message hook, and WritRun's section of `AGENTS.md` — the rest
  of that file stays exactly as the project wrote it. A
  `writrun:begin`/`writrun:end` section left by a kit before `v0.0.04`
  is the one cut where both are there.
- **Recognises the kit's files outside `.writrun/` by the `writrun-`
  prefix they carry**, so a workflow the project wrote is never removed
  and one a later tag added always is.
- **Never touches `work/`.** Tasks, specs, and reports are the
  project's record, not the kit's — the queue and its history survive
  the tooling that managed them.
- Never touches the project's own docs. What the methodology helped
  write belongs to the repository, not to the kit.
- Shows what will be removed and what will stay, then asks
  ([rules](../rules.md)).

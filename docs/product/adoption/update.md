# `writrun update`

Refreshes an adopted kit to a newer WritRun tag.

- **Writes what the fetched tag ships**, minus the paths the adopter
  owns. A tag that adds a file needs no new release of this binary to
  install it ([coupling](../../technical/engineering/coupling.md)).
- **Never touches the adopter's own paths**: `AGENTS.md`, `CLAUDE.md`,
  `.writrun/settings.json`, `.writrun/gates.md`,
  `.writrun/conventions/`, `docs/` and `work/`. The one exception is
  `docs/writrun-instructions.md`, which the kit installed and the kit
  owns.
- **Seeds `.writrun/gates.md` where it is absent**, and leaves every
  word of it where it is not.
- Removes the kit's own files the tag stopped shipping, and only those:
  outside `.writrun/` it recognises them by the `writrun-` prefix they
  carry.
- Names a `writrun:begin`/`writrun:end` section left in `AGENTS.md` by a
  kit before `v0.0.04`. Cutting it is the adopter's — the file is
  theirs, whole.
- Shows what will change before changing it.

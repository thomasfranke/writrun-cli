# `writrun update`

Refreshes an adopted kit to a newer WritRun tag.

- **Writes what the fetched tag ships**, minus the paths the adopter
  owns. A tag that adds a file needs no new release of this binary to
  install it ([coupling](../../technical/engineering/coupling.md)).
- **Never touches the project's home**: `writrun/` entire, and with it
  `AGENTS.md`, `CLAUDE.md`, `docs/` and `work/`. The one exception is
  `docs/writrun-instructions.md`, which the kit installed and the kit
  owns.
- **Seeds nothing.** A file the project never wrote is answered by the
  kit's own default, so there is nothing for a refresh to put in its
  place — writing into the project's home is adoption's act.
- **Moves the adopter's files once**, where a repository adopted before
  the two homes still keeps them inside `.writrun/`. The move runs
  before anything is written, and the content crosses byte for byte.
- Where both addresses hold the file, the one in the project's home
  stands, the old one is named and left where it is, and nothing is
  merged: two answers about one setting are the adopter's to
  reconcile.
- Removes the kit's own files the tag stopped shipping, and only those:
  outside `.writrun/` it recognises them by the `writrun-` prefix they
  carry.
- Names a `writrun:begin`/`writrun:end` section left in `AGENTS.md` by a
  kit before `v0.0.04`. Cutting it is the adopter's — the file is
  theirs, whole.
- Shows what will change before changing it.

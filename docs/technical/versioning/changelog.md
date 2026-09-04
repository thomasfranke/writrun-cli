# Changelog

- **The history lands in the repository, not only on the forge.** The
  cut writes [`CHANGELOG.md`](../../../CHANGELOG.md) at the root,
  newest first, and commits it; the tag lands on that commit. Entries
  are the conventional commit subjects between the previous tag and
  `HEAD`, read with `git` — the same material the forge's generated
  notes use.
- **`CHANGELOG.md` is generated, never edited by hand** — a second
  writer would fork the history the tags carry. A wrong entry is fixed
  in the commit subject that produced it, on the next tag. It is not a
  permanent doc: no spec promises it, and nothing under `docs/` changes
  when it is written.

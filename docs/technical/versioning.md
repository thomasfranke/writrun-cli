# Versioning

- The version number is computed, never typed: `make release` (default
  `minor`; also `major`, `epoch`) computes the next number from the
  latest tag, writes the changelog, runs the suite, then commits, tags,
  pushes, and publishes the GitHub Release with generated notes. The
  whole path lives in [`scripts/release.sh`](../../scripts/release.sh).
- `minor` bumps the third digit, `major` the middle one, `epoch` the
  first — WritRun's own vocabulary, not SemVer's.
- **The first release is `v0.0.01`, and the third field stays two
  digits.** Same scheme as WritRun's own tags, deliberately: this client
  pins one of them and reads it back, so a client numbering itself
  differently from what it pins reads as two schemes where there is one.
- **The tag is the only place the number is written.** No `VERSION`
  file exists: a stamp is a second copy of what the tag already says,
  wrong the moment a cut fails between the two writes.
- **The history lands in the repository, not only on the forge.** The
  cut writes [`CHANGELOG.md`](../../CHANGELOG.md) at the root, newest
  first, and commits it — the release commit the tag then lands on, so
  the number and what earned it travel together. The entries are the
  conventional subjects between
  the last tag and `HEAD`, read with `git` — the same material the
  forge's generated notes read; what changes is that a checkout answers
  "what changed" without leaving it.
- **It is generated, and never edited by hand.** One writer is what
  keeps the file from becoming a second history that agrees with the
  tags until the first time somebody forgets. An entry that is wrong is
  wrong in the subject that produced it, and that is where it is fixed,
  on the next tag. It is not a permanent doc: no spec promises it, and
  nothing under `docs/` changes when it is written.
- **Makefile** — thin aliases only; the scripts are the interface and
  CI calls them directly. Release: `make release [minor|major|epoch]`.

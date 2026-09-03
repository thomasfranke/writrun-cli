# Versioning

- The version number is computed, never typed: `make release` (default
  `minor`; also `major`, `epoch`) stamps the root `VERSION`, runs the
  suite, then commits, tags, pushes, and publishes the GitHub Release
  with generated notes. The whole path lives in
  [`scripts/release.sh`](../../scripts/release.sh).
- `minor` bumps the third digit, `major` the middle one, `epoch` the
  first — WritRun's own vocabulary, not SemVer's.
- **The first release is `v0.0.01`, and the third field stays two
  digits.** Same scheme as WritRun's own tags, deliberately: this client
  pins one of them and reads it back, so a client numbering itself
  differently from what it pins reads as two schemes where there is one.
- `.writrun/VERSION` is the adopted WritRun kit's version, never this
  project's. The root `VERSION` is this project's.
- **Makefile** — thin aliases only; the scripts are the interface and
  CI calls them directly. Release: `make release [minor|major|epoch]`.

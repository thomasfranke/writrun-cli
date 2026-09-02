# Versioning

- The version number is computed, never typed: `make release` (default
  `minor`; also `major`, `epoch`) stamps the root `VERSION`, runs the
  suite, then commits, tags, pushes, and publishes the GitHub Release
  with generated notes. The whole path lives in
  [`scripts/release.sh`](../../scripts/release.sh).
- `minor` bumps the third digit, `major` the middle one, `epoch` the
  first — WritRun's own vocabulary, not SemVer's.
- `.writrun/VERSION` is the adopted WritRun kit's version, never this
  project's. The root `VERSION` is this project's.
- **Makefile** — thin aliases only; the scripts are the interface and
  CI calls them directly. Release: `make release [minor|major|epoch]`.

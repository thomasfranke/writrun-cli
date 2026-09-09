# Release

- The version number is computed, never typed: `make release` (default
  `patch`; also `minor`, `major`) computes the next number from the
  latest tag, writes the changelog, runs the suite, then commits, tags,
  pushes, and publishes the GitHub Release with generated notes. The
  whole path lives in
  [`scripts/release.sh`](../../../scripts/release.sh).
- **With no tag yet, the bump still decides**: `patch` cuts
  [`v0.0.1`](scheme.md), `minor` cuts `v0.1.0`, `major` cuts `v1.0.0`.
  A first cut that answered the same number whatever it was asked for
  would be a command that does not mean what it says.

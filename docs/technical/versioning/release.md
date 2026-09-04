# Release

- The version number is computed, never typed: `make release` (default
  `patch`; also `minor`, `major`) computes the next number from the
  latest tag, writes the changelog, runs the suite, then commits, tags,
  pushes, and publishes the GitHub Release with generated notes. The
  whole path lives in
  [`scripts/release.sh`](../../../scripts/release.sh).

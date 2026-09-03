# Changelog

Notable changes, curated by hand. The machine record is elsewhere: every
tagged release publishes GitHub notes generated from its commits, and
`git log` is complete. This file is the shorter read — what changed for
someone using `writrun`, not every commit that got there.

The numbering is WritRun's own vocabulary, not SemVer's: `minor` bumps the
third digit, `major` the middle one, `epoch` the first. The number is
computed by `make release`, never typed — see
[`versioning.md`](docs/technical/versioning.md).

## Unreleased

Nothing has been tagged yet. The root [`VERSION`](VERSION) reads `v0.0.0`,
which is the state before the first cut; the first release is `v0.0.1`.

### Added

- The release path — `scripts/release.sh` and `make release`, computing the
  next number from the latest tag, stamping `VERSION`, running the suite,
  then committing, tagging, pushing and publishing the GitHub Release.
- The test suite — tiers under `tests/`, twelve cases covering the release
  path, on git and POSIX shell alone.
- The documentation that defines the binary:
  [`docs/product/`](docs/product/README.md) command by command, and
  [`docs/technical/`](docs/technical/README.md) for how it is built,
  versioned and distributed.
- WritRun adopted — the kit under [`.writrun/`](.writrun/README.md), the
  four workflows, `AGENTS.md`, and the `work/` queue. This project runs on
  the methodology it is a client for.

The binary itself is not implemented. `docs/product/` names every command;
the queue in `work/` is the gap between those docs and the code.

---
id: task-0032
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0035]
doc_ref: technical/versioning/scheme.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-09T07:40:00Z
queued: 2026-09-09T09:05:44Z
completed: null
merged: null
provenance: []
---

# Cut the first release as the bump asks, starting at v0.0.1

`scripts/release.sh` answers `v0.1.0` for the first cut, hardcoded, and
discards the `patch|minor|major` it was given. The maintainer ran a cut
expecting `v0.0.1` and got `v0.1.0` with no word about why.

Two things are wrong and only one is about the number:

- **The number.** The rule now says the first release is `v0.0.1` — as
  far down as SemVer allows, which is where WritRun's own numbering
  starts. `v0.0.01`, what upstream actually cuts, is not SemVer:
  decision
  [0008](../../docs/technical/decisions/versioning/0008-the-cli-version-is-semver.md)
  records that the leading zero makes the Go module system ignore the
  tag and GoReleaser refuse the cut.
- **The argument.** Whatever the first number is, a cut that is asked
  for `patch` and answers something else without saying so is a command
  that does not mean what it says. The rule now states what each bump
  cuts from nothing.

No tag exists yet in this repository, so this is fixed before it has
ever been lived with — nothing published needs renumbering.

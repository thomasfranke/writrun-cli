# Scheme

- **The version is SemVer** — `vMAJOR.MINOR.PATCH`, the form the Go
  module system, GoReleaser and Homebrew all parse; WritRun's own
  scheme stays upstream's (decision
  [0008](../decisions/versioning/0008-the-cli-version-is-semver.md)).
- **The first release is `v0.1.0`.**
- **The tag is the only place the number is written.** No `VERSION`
  file exists: a stamp is a second copy of what the tag already says,
  wrong the moment a cut fails between the two writes.

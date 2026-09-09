# Scheme

- **The version is SemVer** — `vMAJOR.MINOR.PATCH`, the form the Go
  module system, GoReleaser and Homebrew all parse; WritRun's own
  scheme stays upstream's (decision
  [0008](../decisions/versioning/0008-the-cli-version-is-semver.md)).
- **The first release is `v0.0.1`.** It starts where WritRun's own
  numbering starts, as far down as SemVer allows — `v0.0.01` is what
  upstream cuts and is not SemVer at all, which is the whole of
  decision 0008. A leading zero is not decoration here: it is what
  makes the Go module system ignore the tag.
- **The tag is the only place the number is written.** No `VERSION`
  file exists: a stamp is a second copy of what the tag already says,
  wrong the moment a cut fails between the two writes.

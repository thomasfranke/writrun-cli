---
id: spec-0012
task_ref: task-0013
status: implemented
created: 2026-09-03T22:30:44Z
---

# spec-0012 — GoReleaser builds, the tap follows, go install works

**References:** [task-0013](../tasks/task-0013-release-distribution.md)

- **Goal:** a tag on `main` becomes installable three ways without a manual step.

## Scope

In: GoReleaser configuration, the tap formula, release artifacts,
`go install` compatibility, the WritRun tag named in each release. Out:
the release cut itself (`scripts/release.sh`, already shipped).

## Steps

1. `.goreleaser.yaml`: static builds for macOS, Linux, Windows; no cgo; version stamped from the tag.
2. Tap: formula `writrun` in `thomasfranke/homebrew-tap`, updated by GoReleaser on each release.
3. Release notes template names the pinned WritRun tag.
4. Wire GoReleaser into the tagged-release path.

## Acceptance criteria (EARS)

- When a tag lands on `main`, prebuilt binaries for every supported platform shall be published on the GitHub release.
- When a release is published, `brew install thomasfranke/tap/writrun` shall install that version.
- When a release is published, its notes shall name the WritRun tag it targets.
- When run against the module path, `go install github.com/thomasfranke/writrun-cli/cmd/writrun@latest` shall build.

## Edge cases

- A tag cut before the Go code exists: the pipeline must fail loudly, not publish an empty release.

## Tests required

A dry-run GoReleaser check in CI; the release path exercised on a
disposable tag before first use.

## Definition of Done

- [ ] One real release installable by all three routes.
- [x] Suite green.

## Proposed product changes

- none.

## Proposed technical changes

- none — `technical/distribution/routes.md` already states the three routes.

## Outcome

`.goreleaser.yaml` builds `cmd/writrun` for macOS, Linux and Windows on
amd64 and arm64, without cgo, stamping the tag into `main.version` —
the seam `buildVersion()` reads before falling back to the module
version `go install` records, so all three routes name the same number.

`.github/workflows/release.yml` runs GoReleaser on `refs/tags/v*`. The
cut stays `scripts/release.sh`'s: it creates the GitHub release with
generated notes, and GoReleaser appends the platform archives, the
checksums, and a footer naming the pinned WritRun tag, then writes
`writrun` into `thomasfranke/homebrew-tap`.
`scripts/writrun-tag.sh` reads that tag from `cmd/writrun/main.go`, the
one place it is written, and exits 1 when it is absent — a release that
cannot name its tag is not published. Both existing workflows install
goreleaser before `make tests`, which is where the dry-run check runs.

The tap artefact is a cask, not a formula: GoReleaser deprecated
formula generation and `goreleaser check` exits 2 on a configuration
asking for one (report-0008). `brew install thomasfranke/tap/writrun`
resolves either way.

Verified without releasing: `goreleaser check` exits 0;
`goreleaser build --snapshot --clean` links all six binaries and the
snapshot printed `writrun v0.0.0 (pins WritRun v0.0.03)`. Four
integration cases and one e2e case under `tests/*/release/` hold the
matrix, the linked stamp, the footer and the tap.

Two things only a tag can prove, and the first Definition-of-Done box
waits on them: that the release job publishes the archives, and that
the tap push succeeds — the latter needs a `HOMEBREW_TAP_TOKEN`
repository secret, which lives outside the repository and does not
exist yet.

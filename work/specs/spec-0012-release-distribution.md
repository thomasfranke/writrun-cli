---
id: spec-0012
task_ref: task-0013
status: draft
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
- [ ] Suite green.

## Proposed product changes

- none.

## Proposed technical changes

- none — `technical/distribution/routes.md` already states the three routes.

## Outcome

_(fill after execution)_

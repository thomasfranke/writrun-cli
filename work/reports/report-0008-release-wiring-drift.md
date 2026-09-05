---
id: report-0008
status: fixed
task_ref: []
doc_ref: technical/distribution/routes.md
created: 2026-09-05T18:19:25Z
triaged: 2026-09-05T19:03:11Z
---

# Distribution and CI docs are behind the release wiring

**References:** [technical/distribution/routes.md](../../docs/technical/distribution/routes.md)

`technical/distribution/routes.md` calls the tap artefact a formula, and
`technical/testing/ci.md` opens with "Two workflows". Implementing
spec-0012 contradicted both statements. GoReleaser deprecated `brews`:
`goreleaser check` exits 2 on a configuration that asks for a formula
(v2.18.0), so `.goreleaser.yaml` declares `homebrew_casks` and the tap
receives `Casks/writrun.rb` — `brew install thomasfranke/tap/writrun`
still resolves, but "formula" no longer names what is published. The
tagged-release path needed a workflow of its own,
`.github/workflows/release.yml`, so this repository's own workflows are
three. spec-0012 proposed no technical documentation changes, so both
sentences were left as they stand.

**Fixed here.** `routes.md` now says cask and states why GoReleaser
leaves no choice; `ci.md` now opens on three workflows and names
`release.yml`, adding that the `writrun-*` workflows beside them are the
kit's rather than this project's.

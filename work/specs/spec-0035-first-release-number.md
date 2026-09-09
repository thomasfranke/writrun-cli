---
id: spec-0035
task_ref: task-0032
status: draft
created: 2026-09-09T07:40:00Z
---

# spec-0035 — The first cut answers the bump it was asked for

**References:** [task-0032](../tasks/task-0032-first-release-number.md)

- **Goal:** `make release` with no tag in the repository cuts the number
  the bump names, and the default cuts `v0.0.1`.

## Scope

In: how `scripts/release.sh` computes the number when no tag exists.

Out: the number's computation from an existing tag, which is correct
and untouched; the changelog, the suite, the commit, the tag, the push
and the publish, none of which this reaches. Out too: WritRun's own
scheme — `v0.0.01` is not SemVer, and decision
[0008](../../docs/technical/decisions/versioning/0008-the-cli-version-is-semver.md)
stands.

## Steps

1. Compute the first number from the bump rather than answering a
   constant: `patch` gives `v0.0.1`, `minor` gives `v0.1.0`, `major`
   gives `v1.0.0`.
2. Say it in the line the script already prints, so a reader sees which
   bump produced which number before anything is mutated.

## Acceptance criteria (EARS)

- When `make release` runs in a repository with no `v*` tag and no
  argument, the system shall cut `v0.0.1`.
- When `make release patch` runs with no tag, the system shall cut
  `v0.0.1`.
- When `make release minor` runs with no tag, the system shall cut
  `v0.1.0`.
- When `make release major` runs with no tag, the system shall cut
  `v1.0.0`.
- When a `v*` tag exists, the system shall compute the next number
  exactly as it does today.
- When the number is computed, the system shall print it with the bump
  that produced it, before anything is mutated.

## Edge cases

- **A tag that is not a version.** Unchanged: the script already asks
  git for `v*` sorted by version, so a tag outside that shape is not
  the latest.
- **A repository whose first cut is a `major`.** `v1.0.0` is what the
  bump names, and the rule says so; nothing here treats the first cut
  as special beyond having no tag to count from.

## Tests required

- Integration, `tests/integration/release/`: the three bumps from no
  tag, each answering its own number, beside the cases that already
  cover bumping from an existing tag.
- Integration, `tests/integration/release/`: the printed line names the
  number and the bump.

## Definition of Done

- [ ] No number is written in the script that the bump did not produce.
- [ ] `scheme.md` and the script agree on what the first release is.
- [ ] The existing bump-from-a-tag cases pass untouched, because that
      path is not what this changes.

## Proposed technical changes

- None. `technical/versioning/scheme.md` and
  `technical/versioning/release.md` carry the rule and are changed in
  the authoring pull request this spec arrives in.

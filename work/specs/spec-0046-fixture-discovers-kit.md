---
id: spec-0046
task_ref: task-0038
status: approved
created: 2026-09-19T14:23:04Z
---

# spec-0046 — A fixture discovers the kit it stages, and nothing may list it

**References:** [task-0038](../tasks/task-0038-fixture-discovers-kit.md)

- **Goal:** `tests/list_lib.sh` stages the kit by copying its trees, and a
  test refuses any fixture that goes back to naming kit files one by one.

## Scope

**In:** `tests/list_lib.sh`; one new guard in `internal/kit`, beside
`coupling_test.go`, which already polices this repository's own source
for a rule of the same shape; the paragraph in
`docs/technical/testing/suites.md` that states the rule.

**Out:** the ten fixtures that already copy trees — they are the shape
this change makes universal, not work. The kit's own layout: a script
sourcing its neighbours is the kit's design and not a defect, which is
why report-0047 is this repository's and not WritRun's.

## Steps

1. In `tests/list_lib.sh`, replace the two `cp "$REPO_ROOT/$VAR"` lines
   with `cp -R` of `.writrun/scripts` and `.writrun/skills`, the two
   lines `doctor_lib.sh` now carries. Keep `LISTER` — the cases name it
   to run it — and drop `QUEUE_LIB`, which exists only as a staging
   list. Rewrite the comment above it: it currently explains why the
   second file is there, and that reason is gone.
2. Adjust the `mkdir -p` above the copies: the two leaf directories it
   creates are made by the tree copies.
3. Add `TestFixturesDiscoverTheKit` to `internal/kit` (package
   `kit_test`). It reads every `tests/*_lib.sh`, and fails naming the
   fixture and the line for any copy of a single file under `.writrun/`
   — a `cp` whose source is a `.writrun/...` path, directly or through
   a variable the same file assigns such a path to — while a `cp -R` of
   a `.writrun/` directory passes.
4. State the rule in `docs/technical/testing/suites.md`, under the
   sentence that already says no file lists the suites and none lists
   the fixtures: what a fixture stages is discovered the same way, and
   the guard is what holds it.

## Acceptance criteria (EARS)

- When a fixture copies a single file whose path is under `.writrun/`,
  the system shall fail the unit tier naming that fixture and the line.
- When a fixture copies a `.writrun/` directory with `cp -R`, the system
  shall not fail on that line.
- When a fixture names a kit path for a reason other than staging it —
  a case running the script by address — the system shall not fail.
- When the list suite runs, the system shall pass every case it passed
  before this change, against a target staged from the kit trees.

## Edge cases

- **A variable holding the path.** `list_lib.sh` and `doctor_lib.sh`
  both assign `.writrun/...` to a name and copy through it, so a check
  that only reads the `cp` line's literal text sees nothing. The
  assignments in the same file are what the check resolves.
- **A case that runs a kit script by address.** `a_stage_raise_is_
  previewed_test.sh` names `$SETTINGS_CHECK` to run it. Naming is not
  staging, and only `cp` of a single file is refused.
- **`tests/init_lib.sh` writes a stub `read_setting.sh`.** It is
  authoring a fake kit for the adoption cases, not staging this
  repository's, and it writes rather than copies — so it is outside what
  the check reads, and the check must not widen to `cat >`.
- **The guard's own absence.** It is held against it: `list_lib.sh` in
  its current shape is the input that must fail it by name before the
  fix lands.

## Tests required

- `TestFixturesDiscoverTheKit`, run against `list_lib.sh` as it stands
  today — it must fail, naming that file — and again after step 1, where
  it must pass across all eleven fixtures.
- `make test-list`, green, with the fixture staging trees.
- `make test-integration` and `make test-e2e`, green — the tree copies
  put more files in the target than the two the lister cases had, and
  nothing may count them.

## Definition of Done

- [ ] `tests/list_lib.sh` copies `.writrun/scripts` and `.writrun/skills`
      as trees and names no kit file for staging
- [ ] `TestFixturesDiscoverTheKit` exists, and was seen to fail on the
      pre-change fixture by name
- [ ] `docs/technical/testing/suites.md` states the rule
- [ ] the nine steps of `tests.yml` are green
- [ ] report-0047's finding is answered in the spec's Outcome

## Proposed product changes

- none — the binary's behaviour does not change

## Proposed technical changes

- `technical/testing/suites.md` — a fixture discovers the kit it stages;
  it never lists it, and a test refuses the list.

## Outcome

_(fill after execution)_

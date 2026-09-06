---
id: spec-0018
task_ref: task-0019
status: implemented
created: 2026-09-05T19:11:34Z
---

# spec-0018 — Consolidate the parallel round's duplication

**References:** [task-0019](../tasks/task-0019-consolidate-duplication.md)

- **Goal:** the duplication five parallel tasks took on deliberately is settled now that all five have landed.

## Scope

In: the two duplications the round recorded — reading `.writrun/VERSION`
and comparing tags, and the stage-0/stage-1 checks. Out: any behaviour
change; any command's output; `.writrun/` itself.

## What was duplicated, and why

Both were taken knowingly, because five tasks were in flight on the same
packages and a shared helper each would have had to agree on before any
could land was a coupling none of them could pay for. Now they have all
landed, and the reasons are spent.

- **`.writrun/VERSION`** ([[report-0010]]): `updatecmd` reads it in
  `recordedTag` and orders tags in `compareTags`/`parseTag`; `statuscmd`
  reads the same file with its own `recordedTag`/`sameRelease`/`numbers`
  and compares for equality rather than order; `doctorcmd` parses it a
  third time. One file has three readers, and one comparison has two
  implementations that already disagree on what they compute.
- **The stage checks** ([[report-0012]]): `initcmd/checks.go` and
  `doctorcmd/files.go` each carry the stage-0 PATH probe and the
  stage-1 file, marker and VERSION checks. They already disagree on
  three points — where the declared stage comes from, whether
  `check_front_matter.sh` and `check_settings.sh` run, and how the four
  human gates are tested.

## Steps

1. Decide, per duplication, whether it is extracted or kept — a duplication that is cheaper than the coupling is a valid answer, recorded rather than assumed.
2. Where extracted, name the interface where it is consumed and put a fake beside it if it leaves the process (`technical/engineering/boundaries.md`).
3. Where the copies disagree, say which behaviour is correct before unifying — the three disagreements above are behaviour, not formatting.
4. Leave every command's output byte-identical; this is a refactor.

## Acceptance criteria (EARS)

- When the change is complete, `.writrun/VERSION` shall have one reader or a recorded reason for more than one.
- When a duplication is kept, the reason shall be written down where the next reader will find it.
- When the copies disagreed, the resolution shall name which behaviour won.
- When the suite runs, every existing case shall pass with no edit to the case.

## Edge cases

- `updatecmd` orders tags; `statuscmd` only needs equality. A shared helper must not force ordering semantics on a caller that does not want them.
- `initcmd` reports gaps without blocking adoption; `doctorcmd` grades findings and sets the exit status. Shared checks must not hand either one the other's verdict.

## Tests required

The existing suites, unmodified — this changes no behaviour. Unit tests
for whatever is extracted, matching the density of the packages it came
from.

## Definition of Done

- [x] Each duplication is either extracted or kept with the reason recorded.
- [x] No command's output changed.
- [x] `tests/integration/` passes with no edit to any case.
- [x] Suite green.

## Proposed product changes

- none — no behaviour change.

## Proposed technical changes

- `technical/layout/tree.md` — a row for any package this extracts.

## Outcome

One duplication was extracted, one was kept. No command's output
changed: the three commands were run over the same repository before
and after, with `.writrun/VERSION` recorded, empty, absent, unreadable,
spelled with a leading zero, and ahead of the pinned tag, and every
byte of stdout, stderr and every exit code matched.

**`.writrun/VERSION` — extracted.** `internal/kittag` is the file's one
reader: `Path`, `Read`, `Compare`, `SameRelease`, `Components`,
`Readable`. It reads through the `vfs.FS` port its callers already own,
so it leaves the process through nothing of its own and needs no fake.

- Ordering and equality are separate entry points. `SameRelease` is not
  `Compare(a, b) == 0`: ordering carries `update`'s rule that an
  unreadable tag is a move forward, and `status` inherits none of it.
- `doctor`'s third parse is `Readable`, kept stricter than `Components`
  on purpose. A recorded tag needs a leading `v` and two all-digit
  components; a comparison takes what it can order, because refusing a
  spelling would call a refresh a downgrade.
- The words a bad tag gets stay with each command. `Read` returns the
  empty string and the filesystem's own error and grades nothing, so
  `update` still refuses a refresh, `status` still names the file, and
  `doctor` still writes a finding.

**The stage checks — stage 0 extracted, stage 1 kept.** The stage-0
probe is one list and one loop, identical in both copies and fixed by
`docs/technical/runtime/requirements.md`; it is now
`internal/requirements`, which returns the binaries that are missing
and grades nothing. Stage 1 is not shared, and the reason is at
`initcmd.checkFiles`: `init` reports on the adoption it has just run,
`doctor` reports on a repository in use, and one function answering
both would answer neither.

Which behaviour won, on the three the copies disagree about:

| Disagreement | Correct |
|---|---|
| Where the declared stage comes from | Both. `init` has only the flag it just asked for; `doctor` reads `read_setting.sh` because the file is the repository's answer by then. |
| Whether `check_front_matter.sh` and `check_settings.sh` run | Both. A verdict on the queue is the repository's own script, which is `doctor`'s work; `init` runs no script against a repository still being adopted, and blocks nothing. |
| How the four human gates are tested | `doctor`'s. It names the gate that went unanswered and survives a retitled section, where `init` tests `AGENTS.md` for `<!-- TODO`. `init` asks the adoption-local question and its output is unchanged here; `doctor`'s is the behaviour that wins if the two are ever unified. |

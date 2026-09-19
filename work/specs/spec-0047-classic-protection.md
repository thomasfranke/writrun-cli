---
id: spec-0047
task_ref: task-0039
status: approved
created: 2026-09-19T21:42:51Z
---

# spec-0047 — Every rule over main is read, and the row says what it read

**References:** [task-0039](../tasks/task-0039-classic-protection.md)

- **Goal:** stage 2 reads every rule the forge holds over `main` — the rules rulesets contribute and the classic branch protection rule — and the governance row states one fact instead of two.

## Scope

In: `mainReachable` and the helpers it owns in
`internal/command/doctorcmd/forge.go`; the `mainGoverned` requirement's
name and its `advises` note; the stage-2 sentence and the two `main`
rows of the table in `product/adoption/doctor.md`, and the copy of that
table in `internal/command/doctorcmd/explanations.go`; the two frames
that draw the rows, `product/screens/adoption/doctor.excalidraw` and
`product/screens/adoption/config.excalidraw`; the Go and integration
cases that stub the forge reads.

Out: `canWrite` and every stage-0, stage-1 and stage-3 check. Out: the
four rules `blockers` names, which this spec maps onto rather than
grows. Out: which bypass actor the forge resolves the Actions token to,
and what it offers on a user-owned repository — spec-0024's premise,
unchanged.

**The load-bearing premise:** `rules/branches/main` answers with the
rules rulesets contribute and nothing else, so a classic branch
protection rule is invisible to it.
[report-0048](../reports/report-0048-doctor-reads-only.md) is the
evidence: `main` carried a classic rule, the endpoint reported no rule,
and doctor said nothing governed the branch. If that is wrong, this
spec has no subject.

## What is wrong

`mainReachable` makes one read and writes two rows from it. A branch a
classic rule governs therefore gets both answers inverted at once:

- `main is governed by a ruleset` is graded `advises`, though a rule
  does govern the branch;
- `no rule over main refuses the recording push` is graded `met`,
  though `required_pull_request_reviews` refuses every direct push the
  Actions bot makes;
- the recording push is refused on the first merge, after doctor
  reported stage 2 clear.

The governance row states and denies one fact in one sentence:
`main is governed by a ruleset — no ruleset governs it — the
methodology recommends protecting it`. `requirement.text()` renders
`name — note`, the glyph already carries the denial, and this is the
only row whose note repeats it.

## Steps

1. Ask whether a classic rule exists before reading one:
   `repos/{owner}/{repo}/branches/main --jq .protection.enabled`. The
   branch object answers `200` whether or not the branch is protected,
   so nothing is inferred from a `404`. `.protected` cannot carry this
   — a branch governed by a ruleset alone answers `.protected: true`
   with `.protection.enabled: false`.
2. Where it is `true`, read `repos/{owner}/{repo}/branches/main/protection`
   and map its fields onto the vocabulary `blockers` already states:

   | The classic rule carries | The blocker it is |
   |---|---|
   | `required_pull_request_reviews` | require a pull request before merging |
   | `required_status_checks` | require status checks to pass |
   | `required_signatures` enabled | require signed commits |
   | `restrictions`, unless `restrictions.apps` names `github-actions` | restrict updates |

3. Make governance the question of whether anything governs `main`:
   either read answering yes is `met`. Rename the requirement to
   `main is governed by a protection rule`, one term for both kinds.
4. Drop the denial from the `advises` note, leaving the advice alone:
   `the methodology recommends protecting it; nothing blocks the
   recording push meanwhile`.
5. Give the classic rule its own refusal sentence. A classic rule names
   no bypass list the Actions bot can be on, and `enforce_admins`
   changes nothing because the bot is never an admin — so the rule is
   named on both owner types and the remedy is taking it off `main`.
6. Amend the stage-2 sentence and the two table rows in `doctor.md`,
   the copy of that table in `explanations.go`, and the two drawn
   frames that carry the row's text.

## Acceptance criteria (EARS)

- When the classic rule over `main` requires a pull request, the system shall report `no rule over main refuses the recording push` as breaking, on a user-owned and an organization-owned repository alike.
- When a classic rule governs `main` and no ruleset does, the system shall report the governance row as met.
- When neither a ruleset nor a classic rule governs `main`, the system shall report the governance row as `advises` with a note that does not deny the row's own name.
- When the classic rule carries only `required_linear_history`, `required_conversation_resolution`, `allow_force_pushes: false` and `allow_deletions: false`, the system shall report no finding — a plain fast-forward push meets none of them.
- When `restrictions.apps` names `github-actions`, the system shall report no finding for `restrictions`.
- When `branches/main` cannot be read, the system shall report both rows unread, never met.
- When run against this repository's live shape, stage 2 shall report no finding.

## Edge cases

- A branch a ruleset and a classic rule both govern: each source that refuses the push earns one line, and the governance row is met once.
- `.protection.enabled` `false`: the protection payload is not read, and the run costs one extra forge read at most.
- A `404` from the protection payload after `.protection.enabled` said `true`: a rule removed mid-run, read as unread rather than as no rule.
- `required_signatures` absent from the protection payload: the sub-endpoint `branches/main/protection/required_signatures` answers it, and the implementation confirms which of the two the forge fills before choosing — the mapping above is the contract, not the endpoint.
- `enforce_admins: false`: it changes no row, because the Actions bot is not an admin.
- A repository whose owner type cannot be read: unchanged from spec-0024, and the classic rule's finding does not ask the question at all.

## Tests required

Go cases over stubbed `gh` answers: a classic rule requiring a pull
request on each owner type, a classic rule carrying only the four rules
a fast-forward meets, `restrictions` with and without `github-actions`,
a branch governed by both kinds, `.protection.enabled` false, and
`branches/main` unreadable. One case asserting the `advises` note no
longer denies its row. One asserting this repository's live shape still
reports nothing.

Every doctor fixture that reaches stage 2 stubs the forge reads by hand;
the new read is owed a `gh_reply` in each, and a fixture missing it is
the failure this change has to not leave behind. Each new case is held
against its absence: the read it guards is broken, and the case named
fails.

The integration cases in `tests/integration/doctor/` that assert the
governance row's text assert the old sentence and must be amended rather
than deleted; the Outcome names each.

## Definition of Done

- [ ] A classic rule requiring a pull request is one breaking finding on both owner types.
- [ ] A classic rule alone makes the governance row met.
- [ ] The `advises` note carries advice and no denial.
- [ ] `doctor.md`, `explanations.go` and the two drawn frames say the same thing.
- [ ] `writrun doctor` still reports no stage-2 finding against this repository.

## Proposed product changes

- `product/adoption/doctor.md` — the stage-2 sentence naming what governs `main`, and the two table rows for the governance and refusal requirements.
- `product/screens/adoption/doctor.excalidraw` — the four frames drawing the governance row's text.
- `product/screens/adoption/config.excalidraw` — the frame drawing the same row in the requirements preview.

## Proposed technical changes

- none — no new package, no new boundary, and the forge reads are stated where the command is documented.

## Outcome

_(fill after execution)_

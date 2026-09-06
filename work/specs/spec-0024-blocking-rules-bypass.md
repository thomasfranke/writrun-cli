---
id: spec-0024
task_ref: task-0025
status: implemented
created: 2026-09-06T07:35:01Z
---

# spec-0024 — Blocking rules judged against their bypass list

**References:** [task-0025](../tasks/task-0025-blocking-rules-bypass.md)

- **Goal:** stage 2 names a ruleset rule only where the recording push cannot get past it — one finding per ruleset that stops the push, and none for a ruleset the Actions bot can bypass.

## Scope

In: `mainReachable`, `blocking`, `named` and `firstOf` in
`internal/command/doctorcmd/forge.go`; the `doctor.md` sentence naming
the four blocking rules; the Go and integration cases that assert the
current behaviour.

Out: which bypass actor the forge resolves the Actions token to, which
stays the forge's answer. Out: the set of four rules, which no evidence
asks to grow or shrink. Out: `writrun doctor`'s output shape, every
stage-0, stage-1 and stage-3 check, and the `canWrite` half of stage 2,
unchanged since spec-0019.

**The load-bearing premise**, from `.github/workflows/writrun-approve.yml`:
the forge offers GitHub Actions as a bypass actor on organization-owned
repositories alone. If that is wrong, this spec's shape is wrong.

## What is wrong

`blocking()` reads the rule types on `main` and never a bypass list, and
`mainReachable` appends `named(on)` after the per-ruleset loop that read
one. Three consequences:

- a ruleset naming a bypass actor and enabling `update`,
  `required_signatures` or `required_status_checks` is reported `breaks`
  although the bot is past the rule;
- `pull_request` is filtered out by `userOnly` on an organization-owned
  repository, so a ruleset requiring a pull request with no bypass actor
  is reported `all clear` — a configuration where the push cannot land;
- a blocking rule with an empty bypass list produces two findings whose
  remedies contradict.

spec-0019 put these four out of scope on the premise that they already
read capability, and its Outcome opens "Stage 2 reads the capability".
Both halves of its goal are unmet: stage 2 fails a configuration where
the push lands, and passes one where it does not.

## Steps

1. Amend `doctor.md`'s stage-2 sentence: a rule that refuses the recording push is named where the ruleset enabling it gives the Actions bot no way past — on a user-owned repository always, on an organization-owned one where the ruleset names no bypass actor. This is an authored rule change and takes the authoring flow.
2. Move the blocking decision inside `mainReachable`'s per-ruleset loop: for each ruleset governing `main`, judge the rules it contributes against that ruleset's own bypass list.
3. Replace `userOnly` with the ownership rule the premise states, `pull_request` included, keeping the ownership read at most once per run.
4. Collapse the two finding shapes into one per ruleset, so a single fault produces a single line.

## Acceptance criteria (EARS)

- When a ruleset names a bypass actor and enables `update` on an organization-owned repository, the system shall report no finding.
- When the same ruleset governs a user-owned repository, the system shall report a finding.
- When `pull_request` is enabled with no bypass actor on an organization-owned repository, the system shall report a finding naming it.
- When a blocking rule is enabled with an empty bypass list, the system shall report exactly one finding.
- When run against this repository's live shape, stage 2 shall report no finding.

## Edge cases

- Two rulesets governing `main`, one bypassable and one not: the finding names the one that stops the push, not the other.
- A ruleset whose bypass list names an actor that is not the Actions bot: which actor the token resolves to stays the forge's answer, and this spec does not guess it.
- `gh` unauthenticated: stage 2 stays unread, as it already does.
- A repository whose owner type cannot be read: it is not a licence to pass silently.

## Tests required

Go cases over stubbed `gh` answers for each cell of ownership × bypass ×
rule, including the two that are wrong today: a bypassed `update` on an
org repository passing, and `pull_request` with an empty bypass list on
an org repository failing. One case asserting a single finding where
there is a single fault. One asserting this repository's live shape
passes.

Four existing Go cases and two integration cases assert the current
behaviour and must be inverted rather than deleted; the Outcome names
each and why it changed.

## Definition of Done

- [x] A bypassable ruleset reports no finding on an organization-owned repository.
- [x] A `pull_request` rule with no bypass actor reports one on an organization-owned repository.
- [x] One fault produces one finding.
- [x] `writrun doctor` still reports no stage-2 finding against this repository.

## Proposed product changes

- `product/adoption/doctor.md` — the stage-2 sentence naming the blocking rules.

## Proposed technical changes

- none — no new package and no new boundary.

## Outcome

Stage 2 judges a rule against the bypass list of the ruleset that
enables it. `mainReachable` decides inside the per-ruleset loop: the
rules one ruleset contributes to `main` say whether it refuses the
recording push, and that same ruleset's bypass list says whether the bot
is past them. `blocking` and `named` are gone, and with them the second
finding; `firstOf` reads `blockers` directly against one ruleset's
rules. `blocker.userOnly` is replaced by `ownership`, which reads
`.owner.type` at most once per run and only where a ruleset refuses the
push — an owner type the forge will not answer is read as a person's, so
the finding stands rather than passing in silence.

One ruleset that stops the push is one finding, and the owner decides
its remedy: an organization is told to put the bot on the ruleset's
bypass list, a person is told to take the rule off `main`. The
`doctor.md` stage-2 sentence names the four rules by the words the
findings use and states when each is named.

The premise held. `.github/workflows/writrun-approve.yml` is the one
place this repository states it, and it states it as the spec quotes it.

Six cases were inverted:

| Case | Why it changed |
|---|---|
| `TestARuleThatRefusesThePushWithNoBypassActorNamesTheRule` | It asserted the contradicting pair. It now asserts one breaking finding for one fault. |
| `TestTheBypassFindingNamesOnlyTheRulesetThatEnablesTheRule` | It asserted the removed `named()` sentence alongside the attribution. The attribution is unchanged; the sentence it checks is the per-ruleset one. |
| `TestTheFourBlockingRulesAreNamedWhenOn` | Its name claimed a rule is named whenever it is on, which is no longer true. Renamed `TestTheFourBlockingRulesAreNamedOnAUserOwnedRepository` and given a ruleset with a bypass actor, which clears nothing there. |
| `TestThePullRequestRuleIsNamedOnlyOnAUserOwnedRepository` | It asserted the false negative — `pull_request` dropped on an organization. Replaced by `TestTheFourBlockingRulesAreNamedOnAnOrganizationWithNoBypassActor`, a table over all four. |
| `the_recording_push_can_reach_main_test.sh`, "a rule that refuses the push with no bypass actor names it" | It expected the organization-shaped sentence on a user-owned fixture. It expects the user-owned one, and a second check proves the same ruleset holds on an organization. |
| `the_forge_settings_are_named_test.sh`, "a rule that blocks the recording push is named" | It expected the removed `named()` sentence. It expects the per-ruleset one, and the pull-request check below it now asserts the finding rather than the silence. |

Four cases were added in `internal/command/doctorcmd/forge_test.go`:
`TestABypassActorClearsARuleOnlyOnAnOrganization` covers ownership ×
bypass on one rule, `TestTheFindingNamesTheRulesetThatStopsThePush`
covers two rulesets of which one is bypassed,
`TestAnUnreadableOwnerTypeStillNamesTheRule` covers the owner type the
forge refuses, and `TestOwnershipIsReadOnceForSeveralRulesets` holds the
read to one. `TestThisRepositoryHasNoStageTwoFinding` is unchanged and
still passes.

`writrun doctor` reports no stage-2 finding against this repository,
before and after: it is user-owned, its one ruleset enables none of the
four rules, and its bypass list is empty.

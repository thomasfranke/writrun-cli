---
id: spec-0024
task_ref: task-0025
status: approved
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

- [ ] A bypassable ruleset reports no finding on an organization-owned repository.
- [ ] A `pull_request` rule with no bypass actor reports one on an organization-owned repository.
- [ ] One fault produces one finding.
- [ ] `writrun doctor` still reports no stage-2 finding against this repository.

## Proposed product changes

- `product/adoption/doctor.md` — the stage-2 sentence naming the blocking rules.

## Proposed technical changes

- none — no new package and no new boundary.

## Outcome

_(fill after execution)_

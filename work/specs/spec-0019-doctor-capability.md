---
id: spec-0019
task_ref: task-0020
status: implemented
created: 2026-09-06T00:37:44Z
---

# spec-0019 — Doctor checks capability, not configuration

**References:** [task-0020](../tasks/task-0020-doctor-capability.md)

- **Goal:** stage 2 passes for any forge configuration in which the recording push can land, and fails for one in which it cannot.

## Scope

In: the two stage-2 checks that read a configuration — workflow
permissions and the ruleset bypass list — and the `doctor.md` sentences
that specify them. Out: the four blocking-rule checks, which already
read capability and stay as they are; every stage-0, stage-1 and stage-3
check; `writrun doctor`'s output shape.

## What is wrong

`doctor.md` requires workflow permissions `read-and-write` and the
Actions bot on the bypass list, giving the reason: "so the recording bot
can push to `main`". Both are one way to reach that end, not the end.
A repository can satisfy the end and fail both checks:

- Workflow permissions default to `read` while each workflow that writes
  raises `contents: write` itself. This is the tighter configuration —
  the workflows that do not write never receive the right.
- The ruleset enables no rule a fast-forward push meets, so there is
  nothing to bypass and an empty bypass list denies nothing.

This repository is that repository, and its bot carries 52 commits on
`main`.

## Steps

1. Amend `doctor.md`'s stage-2 sentences: state the end — the recording push can reach `main` — and let the checks name what would stop it. This is an authored rule change and takes the authoring flow.
2. Replace the workflow-permissions check: read-and-write passes; `read` passes when every workflow that pushes declares `contents: write` of its own; `read` with a pushing workflow that declares nothing is the finding.
3. Replace the bypass check: a bypass actor passes; no bypass actor passes when the ruleset enables no rule that refuses a fast-forward push; no bypass actor with such a rule enabled is the finding, naming the rule.
4. Keep the four blocking-rule checks exactly as they are — they already read capability.

## Acceptance criteria (EARS)

- When workflow permissions are `read` and every pushing workflow declares `contents: write`, stage 2 shall pass without a finding.
- When workflow permissions are `read` and a pushing workflow declares nothing, the system shall report a finding naming that workflow.
- When the ruleset enables no rule that refuses a fast-forward push, an empty bypass list shall not be a finding.
- When a rule that blocks the recording push is enabled and no bypass actor is set, the system shall report a finding naming the rule.
- When run against this repository as it stands, stage 2 shall report no finding.

## Edge cases

- A forge that offers no rulesets at all: the bypass check has nothing to read and must say so rather than pass silently.
- A workflow that pushes only on a branch, never to `main`: it is not a recording workflow and its permissions are not stage 2's business.
- `gh` unauthenticated: stage 2 stays unread, as it already does, rather than reporting findings it could not check.

## Tests required

Integration over stubbed `gh` answers: the tight configuration passing,
the loose configuration passing, a pushing workflow with no declared
permission failing, and a blocking rule with no bypass actor failing.
One case asserting this repository's own live shape passes — the
regression report-0013 recorded.

## Definition of Done

- [x] `writrun doctor` reports no stage-2 finding against this repository.
- [x] A configuration that genuinely blocks the recording push still fails, proven by a test.
- [x] `doctor.md` states the end rather than one means to it.
- [x] Suite green.

## Proposed product changes

- `product/adoption/doctor.md` — the two stage-2 sentences that name a
  configuration are rewritten to name the capability.

## Proposed technical changes

- none.

## Outcome

Stage 2 reads the capability. `canWrite` passes a repository default of
read-and-write outright, and passes a default of `read` where every
workflow that pushes to `main` raises `contents: write` of its own —
naming the workflow file that does not. `mainReachable` reports an
empty bypass list only where the same ruleset enables one of the four
rules that refuse a fast-forward push, and names that rule; the
entry's `ruleset_id` in the forge's `rules/branches/main` answer is
what attributes a rule to a ruleset. The four blocking-rule checks are
unchanged.

`writrun doctor` reports no stage-2 finding against this repository,
where the binary built from `origin/main` reported the two report-0013
recorded. Eight cases were added in
`internal/command/doctorcmd/forge_test.go`, including the live shape of
`.github/workflows` read from disk.

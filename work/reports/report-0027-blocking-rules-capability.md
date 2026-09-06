---
id: report-0027
status: tracked
task_ref: [task-0025]
doc_ref: product/adoption/doctor.md
created: 2026-09-06T04:27:14Z
triaged: 2026-09-06T07:35:00Z
---

# The four blocking-rule checks report a rule a bypass actor already passes

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md) · [task-0025](../tasks/task-0025-blocking-rules-bypass.md)

spec-0019 set out that stage 2 "passes for any forge configuration in
which the recording push can land, and fails for one in which it
cannot", and its step 4 kept the four blocking-rule checks on the
premise that "they already read capability". They do not. `named()` in
`internal/command/doctorcmd/forge.go` reports `update`,
`required_signatures`, `required_status_checks` and `pull_request` as
`breaks` whenever the rule is enabled, without reading the ruleset's
bypass list — and an actor on that list is past
the rule, so the push lands. A repository with such a rule enabled and
the Actions bot on the bypass list therefore fails stage 2 while
satisfying the end stage 2 exists to check, which is report-0013's
finding in a second place.

The gap is recorded rather than closed because spec-0019's Scope puts
those four checks explicitly out — "Out: the four blocking-rule checks,
which already read capability" — so closing it inside task-0020 would
have been work the approved spec declined. The two checks spec-0019 did
replace, workflow permissions and the bypass list itself, now read
capability; these four are what is left of the shape it was written to
end. Noted while implementing task-0020.

**Correction.** This report first named the four rules
`required_status_checks`, `required_pull_request_reviews`,
`required_signatures` and `required_deployments`. Two of those appear
nowhere in this repository. The list `blockers` holds is `update`,
`required_signatures`, `required_status_checks` and `pull_request`, and
the text above is corrected to it.

**Triage found two more, and resized the first.** The false positive is
an **organization-owned** phenomenon only: `.github/workflows/writrun-approve.yml`
states the forge offers GitHub Actions as a bypass actor on
organization-owned repositories alone, so on a user-owned repository —
which this one is — no bypass clears any of the four and the current
reporting is correct.

The mirror defect is worse. `pull_request` is filtered out by
`userOnly`, so an organization-owned repository with that rule on and an
empty bypass list — a configuration in which the push provably cannot
land — is reported `all clear`. That is the other half of spec-0019's
goal, failing in silence.

And on any repository, a blocking rule with an empty bypass list
produces two findings whose remedies contradict: one says to put the bot
on the bypass list, the other fails the repository for the rule being
on. `TestARuleThatRefusesThePushWithNoBypassActorNamesTheRule` asserts
the pair.

Tracked as task-0025.
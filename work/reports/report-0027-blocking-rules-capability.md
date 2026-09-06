---
id: report-0027
status: open
task_ref: []
doc_ref: product/adoption/doctor.md
created: 2026-09-06T04:27:14Z
triaged: null
---

# The four blocking-rule checks report a rule a bypass actor already passes

**References:** [product/adoption/doctor.md](../../docs/product/adoption/doctor.md)

spec-0019 set out that stage 2 "passes for any forge configuration in
which the recording push can land, and fails for one in which it
cannot", and its step 4 kept the four blocking-rule checks on the
premise that "they already read capability". They do not. `named()` in
`internal/command/doctorcmd/forge.go` reports `required_status_checks`,
`required_pull_request_reviews`, `required_signatures` and
`required_deployments` as `breaks` whenever the rule is enabled, without
reading the ruleset's bypass list — and an actor on that list is past
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

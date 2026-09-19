---
id: report-0049
status: routed
task_ref: []
doc_ref: null
created: 2026-09-19T21:45:08Z
triaged: 2026-09-19T21:55:55Z
---

# The kit's routing section names one destination and no shape

`.writrun/AGENTS.md`, *When the defect is WritRun's*, gives a routing
agent one address: "open the issue on the repository this kit came from
— <https://github.com/thomasfranke/writrun>, the provenance pointer
`WRITRUN.md` carries". `WRITRUN.md` carries that same address and
nothing else. A defect in the `writrun` binary has another home, and no
text the kit ships names it.

The same section states the submission's shape in one clause: "the title
states the observation, the body carries the evidence and the tag in
`.writrun/VERSION`". The form the kit ships,
`.github/ISSUE_TEMPLATE/writrun-report.yml`, asks for those same three
fields and is the receiving repository's, not the routing agent's.

Issue #168 on this repository is the evidence, composed by an agent in
an adopter project running WritRun v0.0.09 through `writrun-cli v0.0.3`.
It reached the right repository, which no instruction named. Its first
sentence states the consumer's implementation — "`doctor`'s stage-2
forge checks read only `/repos/{owner}/{repo}/rulesets`" — and that is
not what the code reads; the endpoint is `rules/branches/main`, and the
claim was inferred from the output rather than observed. Its evidence
block reports a classic branch protection rule on `thomasfranke/tom`;
nine minutes after the issue was opened that branch was given a ruleset
and the classic rule removed, so the state the evidence is read against
no longer exists and the issue does not say which state it was read
against.

**Triage — routed.** It became
[thomasfranke/writrun#288](https://github.com/thomasfranke/writrun/issues/288),
labelled `writrun:submitted`, naming the kit tag `v0.0.09`. The subject
is the kit's own `AGENTS.md` and the form it ships; nothing here can fix
either, and the next `writrun update` would overwrite a patch that
tried.

---
id: task-0025
status: done
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0024]
doc_ref: null
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-06T07:34:59Z
queued: 2026-09-06T08:10:19Z
completed: 2026-09-06T08:55:43Z
merged: 2026-09-06T09:10:48Z
provenance: []
---

# Have doctor judge a blocking rule against the bypass list of the ruleset that enables it

**References:** [spec-0024](../specs/spec-0024-blocking-rules-bypass.md)

`writrun doctor` names a ruleset rule as breaking the recording push
without reading the bypass list of the ruleset that enables it. On an
organization-owned repository the Actions bot on that list is past the
rule and the push lands, so the finding is false; spec-0019 left these
four checks alone on the premise that they already read capability, and
they do not.

The mirror is worse and is silent. `pull_request` is filtered out
entirely on an organization-owned repository, so a ruleset requiring a
pull request with no bypass actor — where the push provably cannot land
— is reported `all clear` by the one check whose purpose is to say
otherwise.

And on any repository, one fault produces two findings that contradict
each other's remedy: doctor tells the adopter to put the bot on the
bypass list, then fails them for the rule being on.

This repository is user-owned with no blocking rule enabled, so nothing
here misreports today. It bites an adopter running the arrangement
`writrun-approve.yml` documents as supported.

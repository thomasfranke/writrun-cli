---
id: report-0005
status: fixed
task_ref: []
doc_ref: technical/testing/ci.md#rules
created: 2026-09-05T16:48:21Z
triaged: 2026-09-05T17:16:57Z
---

# release-readiness.yml keeps the moving tags ci.md refuses

**References:** [technical/testing/ci.md#rules](../../docs/technical/testing/ci.md#rules)

`ci.md#rules` states that every action is pinned by commit SHA, and the
rule sits in the doc that names both workflows. `tests.yml` pins its two
actions as of task-0015. `release-readiness.yml` still reads
`actions/checkout@v4` and `actions/setup-go@v5`, so the branch releases
are cut from is verified by actions nobody reviewed. spec-0014 put that
workflow out of scope by name.

**Fixed here.** `release-readiness.yml` now pins both actions to the
commit SHAs `tests.yml` already carries — `actions/checkout` at
`11d5960a` and `actions/setup-go` at `40f1582b`. The two workflows name
the same two SHAs, which is what `ci.md#rules` asks for.

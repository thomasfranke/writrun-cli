---
id: report-0005
status: open
task_ref: []
doc_ref: technical/testing/ci.md#rules
created: 2026-09-05T16:48:21Z
triaged: null
---

# release-readiness.yml keeps the moving tags ci.md refuses

**References:** [technical/testing/ci.md#rules](../../docs/technical/testing/ci.md#rules)

`ci.md#rules` states that every action is pinned by commit SHA, and the
rule sits in the doc that names both workflows. `tests.yml` pins its two
actions as of task-0015. `release-readiness.yml` still reads
`actions/checkout@v4` and `actions/setup-go@v5`, so the branch releases
are cut from is verified by actions nobody reviewed. spec-0014 put that
workflow out of scope by name.

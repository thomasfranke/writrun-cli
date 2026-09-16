---
id: report-0034
status: open
task_ref: []
doc_ref: product/queue/status.md
created: 2026-09-13T21:23:45Z
triaged: null
---

# status names the kit's tag but never the client's own version

**References:** [product/queue/status.md](../../docs/product/queue/status.md)

`writrun status` answers the kit's tag and never the client's own
version. A run in this repository prints `Kit      WritRun v0.0.07 —
the tag this client pins`, and nothing names `writrun-cli v0.0.2`.

Every other surface that says what is installed says both: `--version`
prints the product, its version and the pinned tag, and the entry
screen's header line carries the same three facts. A reader comparing
two machines from `status` alone cannot tell which client produced the
answer, which is the question a version exists to settle.


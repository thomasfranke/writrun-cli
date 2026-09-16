---
id: report-0034
status: tracked
task_ref: [task-0035]
doc_ref: product/queue/status.md
created: 2026-09-13T21:23:45Z
triaged: 2026-09-16T17:05:00Z
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

**Triage — tracked.** [task-0035](../tasks/task-0035-binary-explains.md)
carries it, through [spec-0043](../specs/spec-0043-status-client-version.md):
the `Client` row, read where `--version` reads it. The row is drawn —
`product/screens/tasks/status.excalidraw` has it above `Kit`, and names
this report beside it — so there is no rule left to write, only the
binary to bring to the drawing.

One correction to the observation above: `writrun-cli v0.0.2` is what a
stamped binary answers. `version` is `dev` at `cmd/writrun/main.go` and
is set at the release cut, so a build from source names itself `dev`.
What `status` omits is the client's identity, whichever of the two it
would have printed.

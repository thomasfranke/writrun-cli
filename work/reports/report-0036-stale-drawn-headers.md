---
id: report-0036
status: fixed
task_ref: []
doc_ref: product/screens/README.md
created: 2026-09-13T21:39:45Z
triaged: 2026-09-13T21:55:00Z
---

# two drawings carry a stale version and a stale pinned tag

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

`entry.excalidraw` and `adoption/config.excalidraw` both open their
frame with `writrun-cli v0.1.0 · pins WritRun v0.0.04 · branch main`.

Neither number is this repository's. `cmd/writrun/main.go` pins
`v0.0.07`, `.writrun/VERSION` records `v0.0.07`, and the newest release
tag is `v0.0.2` — there has never been a `v0.1.0`. Both frames were
transcribed from real runs, and both have been overtaken by a release
and a kit bump since.

`screens/README.md` requires a frame to be transcribed from a real run
or built from values read out of this repository's own files. A frame
that was true when drawn and is false now meets neither test, and a
reader checking the binary against the drawing would find the header
wrong in a drawing that is the reference for it.

**Triage — fixed.** Both frames now read `writrun-cli v0.0.2 · pins
WritRun v0.0.07`, read out of `cmd/writrun/main.go` and
`.writrun/VERSION`. No task: the correction is the whole of the work.

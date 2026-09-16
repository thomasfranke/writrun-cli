---
id: report-0039
status: fixed
task_ref: []
doc_ref: product/screens/README.md
created: 2026-09-16T16:46:42Z
triaged: 2026-09-16T16:55:00Z
---

# nine drawings carry the tag the pin left behind

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

Nine drawings under `docs/product/screens/` name WritRun `v0.0.07`, in
thirty-two drawn texts and thirteen distinct sentences.
`cmd/writrun/main.go` pins `v0.0.08` and `.writrun/VERSION` records
`v0.0.08`, both since #138.

The thirteen are not one sentence repeated. Seven frames open with
` writrun-cli v0.0.2 · pins WritRun v0.0.07 · branch main`; `status`
draws ` Kit      WritRun v0.0.07 — the tag this client pins`; `doctor`
draws `  .writrun/VERSION — v0.0.07`; `init` draws
`   source       WritRun v0.0.07 from the pinned release` and
` Adopted WritRun v0.0.07 at stage 1`; `update` draws
` writrun update — WritRun v0.0.06 → v0.0.07` and
` Refreshed to WritRun v0.0.07.`; `first-run` draws
`init        install the kit at v0.0.07 and ask the stage`; `help`
draws ` writrun-cli v0.0.2 (pins WritRun v0.0.07)`. Each names the tag
this binary pins, and each names a tag it no longer does.

`screens/README.md` requires a frame to be transcribed from a real run
or built from values read out of this repository's own files. A frame
that was true when drawn and is false now meets neither test — and
because the drawings are the reference the implementation is checked
against, a reader checking the binary against them would find the
binary right and the drawing wrong.

This is [report-0036](report-0036-stale-drawn-headers.md) a second
time, against the same rule and by the same cause: a frame carries a
value some later change moves, and nothing holds the two together. The
first time it was a release and a kit bump the frames had not caught up
with. This time it was #138, whose own body named the staleness before
it landed.

**Triage — fixed.** All thirty-two now read `v0.0.08`, the tag
`cmd/writrun/main.go` pins and `.writrun/VERSION` records. Nothing else
in the frames moved.

What outlives the correction is that no check holds a drawn value to
the file it was read from. Twice now the same nine files have gone
stale in the same way, and both times a person caught it — the second
time by predicting it, which is not a check either.

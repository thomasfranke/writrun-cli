---
id: report-0041
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-16T19:18:18Z
triaged: 2026-09-16T23:42:44Z
---

# seven specs promise in a form neither gate can read

Seven specs write their Proposed-changes entries as a markdown link
around a backticked path:

    - [`product/queue/status.md`](../../docs/product/queue/status.md) — the
      row and what it answers.

Both gates that read those sections take a bullet only when a backtick
follows it immediately, and neither takes one that opens with `[`.
`check_promise_paths.sh`'s `promised_written` and
`check_deltas.sh`'s `extract_paths` carry the same awk line and the
same sed behind it. So every one of those entries is read as no entry
at all.

spec-0036, spec-0037, spec-0038, spec-0039, spec-0041, spec-0042 and
spec-0043 are the seven; between them they promise nine paths, and the
count either gate reads from them is zero. `check_promise_paths.sh`
answers `No spec this change adds or modifies promises a path —
nothing to judge`, which is a pass, and reads as one.

The consequence lands later and on someone else. `check_deltas.sh`
judges UNDECLARED against the union of what the listed specs promised;
with that union empty, every permanent doc the completing diff touches
is undeclared and the pull request is refused — for naming a document
the spec did name, in a form the checker cannot see. The promise is
also unheld in the other direction: MISSING can never fire, so a spec
may be implemented without the doc it promised ever being written.

The form arrived in 5058894, which rewrote these sections after
`check_promise_paths.sh` refused the Go packages they first named —
the observation [report-0037](report-0037-technical-section-meaning.md)
carries. The paths it wrote are correct; what it added around them is
what the parser cannot read. The template the kit ships shows the bare
form, and says nothing about a link being one.

**Triage — fixed.** The nine entries in the seven specs now open with
the path in backticks and nothing before it, which is the form the
template shows, the form both gates read, and the form WritRun's own
specs use — `kit/`'s repository writes
`- \`product/README.md\` — its links follow the renames` and has never
written a link. Both gates read all nine now; before this change they
read none.

What is lost is the link the form was carrying, and it is worth naming
rather than pretending otherwise: a reader of a spec no longer clicks
through to the chapter it promises. The path is relative to `docs/` and
resolves by hand, which is the trade the kit's own template already
made. The alternative — a backticked path followed by a parenthesised
link — parses, and was rejected as a second spelling of a form that
already has one.

This was urgent rather than tidy. task-0033, task-0034 and task-0035
are `ready` and are carried by six of these seven specs. Whoever took
one would have reached `writrun-check-spec-deltas` at `finish` with an
empty promise set, and been refused UNDECLARED for touching a document
their spec had named.
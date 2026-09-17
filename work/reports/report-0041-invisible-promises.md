---
id: report-0041
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-16T19:18:18Z
triaged: 2026-09-17T13:03:19Z
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

**Triage — fixed.** Two specs still carried the form when this was
read again: spec-0038 and spec-0042. Both are rewritten to the bare
one the kit's template shows, and no bullet under a Proposed-changes
heading in `work/specs/` opens with `[` any more. The other five were
corrected before this report was triaged; what is left of the
observation is the count it gave, and the count is zero.

The gates are not what changed. `extract_paths` and
`promised_written` are the kit's, and teaching either of them to read
a link is a change to a repository this one cannot write
([AGENTS.md](../../AGENTS.md)). The template already shows the form
both can read, so the specs came to it rather than the other way
round.

---
id: report-0037
status: routed
task_ref: []
doc_ref: null
created: 2026-09-13T22:47:21Z
triaged: 2026-09-17T12:15:56Z
---

# the spec template calls the technical section machinery, and the checker calls it docs

`spec.md`, the template WritRun ships, seeds the section with

    ## Proposed technical changes

    - none — no machinery change

and `check_promise_paths.sh` rejects any entry in it that does not
resolve under `docs/`:

    REJECTED: spec-0036 promises `internal/palette`, read as
    docs/internal/palette — `internal` is a repository-root entry and
    docs/internal does not exist, so no diff can ever touch it.

The word the template uses is *machinery*, which is the code; the path
the checker requires is a document under `docs/technical/`. Two shipped
files say two things about the same section.

`writrun-create-task-and-spec`'s SKILL.md does not settle it either: it
asks for "both Proposed-changes sections with real entries" and gives
`path/to/doc.md#anchor` as the example, without saying that the paths
are read relative to `docs/` or that the technical section means
`docs/technical/` rather than the machinery it is named for.

Observed while drafting eight specs against this kit: five of them
named Go packages, and the pipeline refused the branch.

**Triage — routed.** It became
[thomasfranke/writrun#267](https://github.com/thomasfranke/writrun/issues/267),
labelled `writrun:submitted`. The subject is three kit files — the spec
template, `check_promise_paths.sh` and the create-task-and-spec skill —
and nothing here can fix any of them: the next `writrun update`
replaces all three.

Re-checked against `v0.0.08` before it was sent, because this
repository moved its pin in the meantime. The contradiction survives:
the template is byte-identical between the two tags and still seeds the
section with *machinery*, and condition one of the promise gate is
unchanged, at `:240` rather than `:228`. What did change is worth less
comfort, not more — #261 narrowed condition two, so a code path whose
first segment has no repository-root counterpart now passes the early
gate in silence and fails at `writrun-check-spec-deltas` instead, under
a finished branch.

The issue also carries [report-0041](report-0041-invisible-promises.md)
as a second finding, because the two are one shape: the form the gates
accept is stated in one place and not where an author meets it, and a
plausible wrong form fails silently rather than loudly. That one was
this repository's own to fix, and was fixed here.

Nothing is owed locally now. A finding that goes unanswered upstream is
raised again by a second report, never by reopening this one.
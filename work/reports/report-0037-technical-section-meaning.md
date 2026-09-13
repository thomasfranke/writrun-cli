---
id: report-0037
status: open
task_ref: []
doc_ref: null
created: 2026-09-13T22:47:21Z
triaged: null
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

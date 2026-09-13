---
id: report-0038
status: routed
task_ref: []
doc_ref: null
created: 2026-09-13T22:47:24Z
triaged: 2026-09-13T23:05:00Z
---

# doc_ref cannot name a drawing, so a project whose product docs are drawings cannot reference them

`check_front_matter.sh` requires a task's `doc_ref` to be null or a
`.md` path:

    MALFORMED: work/tasks/task-0034-screen-furniture.md: doc_ref
    'product/screens/first-run.excalidraw' is not null or a .md path
    (optionally with #anchor)

This project's product documentation includes `.excalidraw` drawings —
they are the reference the implementation is checked against for
layout, keys and wording, and `docs/product/screens/README.md` states
them as rules. A task derived from a drawing cannot name that drawing in
`doc_ref`; it names the folder's README instead, and the anchor is lost.

The same holds for the Proposed-changes sections, which
`check_promise_paths.sh` reads as chapters: a drawing is a document a
diff can touch, and a promise to update one cannot be expressed.

Observed while deriving three tasks and eight specs from sixteen
drawings.

**Triage — routed.** The shape rule is the kit's, and nothing in this
repository can lift it: a patch here would be overwritten by the next
`writrun update`. Opened upstream as
[thomasfranke/writrun#252](https://github.com/thomasfranke/writrun/issues/252),
widened there past drawings — documentation is not a file format, and a
`.json` stating the shape of a payload is a document by the same test.

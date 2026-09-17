---
id: report-0045
status: open
task_ref: []
doc_ref: product/screens/README.md
created: 2026-09-17T13:06:58Z
triaged: null
---

# two panes name facts the lister's row does not carry

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

Two screen drawings explain a selected task with facts the row they
explain does not hold.
[`tasks/take.excalidraw`](../../docs/product/screens/tasks/take.excalidraw)
reads `task-0002 — available, spec-0002 approved, priority medium`, and
[`tasks/list.excalidraw`](../../docs/product/screens/tasks/list.excalidraw)
reads `task-0024 — available, priority low, no spec`.

The lister writes a row with `printf '  %-10s %-7s %s\n'` — the id, the
priority and the title
(`.writrun/skills/writrun-select-next-task/list_tasks.sh:543`). The
priority is there and the binary says it; the section the row sits
under is there and the binary says that too. Whether a task references
a spec, and whether that spec is approved, is in the task's own front
matter and in the spec's — neither the row nor the option `take` offers
carries either.

Reaching them is a second read of the queue, per row.
`internal/queue` resolves `spec_ref` off a task and `status` off a
spec, and both `internal/command/takecmd` and the screen's wiring in
`cmd/writrun/main.go` would need the filesystem port neither takes
today. task-0034 gave the rows a cursor and a pane and stopped at what
the rows carry, so both frames are asserted by their rows and their
keys (`internal/command/takecmd/frames_test.go`,
`internal/screen/screens_frames_test.go`) and their panes are not.

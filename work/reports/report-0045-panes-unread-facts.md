---
id: report-0045
status: routed
task_ref: []
doc_ref: product/screens/README.md
created: 2026-09-17T13:06:58Z
triaged: 2026-09-17T19:33:23Z
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

**Triage — routed.** It became
[thomasfranke/writrun#270](https://github.com/thomasfranke/writrun/issues/270),
labelled `writrun:submitted`.

**One correction to the observation above, and it is what decided the
route.** It reads as though the facts are simply absent, so that
reaching them costs a second read of the queue. They are not absent:
`list_tasks.sh` already resolves each task's `spec_ref` and each of
those specs' status, because `ready` *is* every spec approved or
implemented — the loops at `:408-412` and `:492-496`. It computes both
and then drops them at the printf, which writes the id, the priority
and the title.

So the second read would be this repository re-doing work the kit has
already done, and `coupling.md` with decision 0013 refuse that outright.
The row carrying what the lister already knows is the kit's to write,
and that is the ask.

The two panes stay unbuilt and their frames stay asserted by their rows
and their keys. Nothing is owed locally: a finding that goes unanswered
upstream is raised again by a second report, never by reopening this
one.
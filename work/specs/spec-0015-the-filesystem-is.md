---
id: spec-0015
task_ref: task-0016
status: draft
created: 2026-09-05T12:28:39Z
---

# spec-0015 — The filesystem is a port, and a fake is how its failures are tested

**References:** [task-0016](../tasks/task-0016-put-the-filesystem.md)

- **Goal:** the filesystem is reached through a port, and its failures are tested by asking the fake to fail.

## Scope

In: a filesystem port and its fake; `initcmd`, `updatecmd`,
`uninstallcmd`, `hook`, `kitfetch` and `wrepo` reaching the filesystem
only through it; production wiring in `cmd/writrun/` alone. In: the
three identical `GitRunner` declarations collapsed to one.

Out — **no behaviour changes**. This is a refactor: every integration
case in `tests/` passes unmodified, and that is the safety property the
work is judged on.

Out: `internal/term`. Its spinner reaches the terminal, not the
filesystem, and its one uncovered statement is a cancellation this task
does not address.

1. Define the port where it is consumed, small enough to name only what
   the commands do: read, write, stat, walk, make and remove a
   directory, and the mode a write keeps.
2. Write the fake beside it: a tree in memory, and a way to say *this
   path's next call fails*, so a test names the failure rather than
   arranging the machine to produce it.
3. Move the six packages onto it, one at a time, with the suite green
   between each.
4. Delete the tests that make a file or a directory read-only, and
   replace each with the fake refusing the call it was arranging.
5. Collapse `kitfetch.GitRunner`, `initcmd.gitRunner` and
   `hook.GitRunner` into one declaration, and drop the conversions in
   `cmd/writrun/main.go`.
6. Record the port and the fake in the layout table, which does not
   name them or the five packages task-0003 and task-0005 extracted
   (report-0003).

## Acceptance criteria (EARS)

- When a package under `internal/command/` reaches the filesystem, it
  shall do so through the port; the package shall import no filesystem
  call from `os` or `path/filepath`'s walkers.
- When the fake is told a path's call fails, the command shall report
  the failure naming that path, and shall leave every other path
  untouched.
- When the suite runs, no test shall change a file's or a directory's
  mode to arrange a failure.
- When the suite runs, coverage over `internal/` shall be at least 99%.
- When the refactor is complete, every case in `tests/integration/`
  shall pass with no edit to the case.
- When the wiring is read, one `GitRunner` type shall be declared, and
  `cmd/writrun/main.go` shall convert between none.

## Edge cases

- `filepath.WalkDir` takes the real filesystem; walking through the port
  means the port owns the walk, or the fake cannot answer it.
- File modes are load-bearing: the commit-message hook is executable,
  and a port that drops the mode installs a hook nobody can run.
- `kitfetch` writes outside the repository, into a temp directory. The
  port reaches there too, or the fetch stays untestable for the same
  reason the writes were.

## Tests required

Unit, per command: the fake refusing each call the command makes, one at
a time, asserting the error names the path and that nothing else was
written. Integration: unchanged — they are the proof the refactor
changed no behaviour.

## Definition of Done

- [ ] No `os` filesystem call remains under `internal/command/`.
- [ ] No test changes a mode to arrange a failure.
- [ ] Coverage over `internal/` is at least 99%.
- [ ] The integration suite passes unmodified.
- [ ] Suite green.

## Proposed product changes

- none — no behaviour change

## Proposed technical changes

- `technical/layout/tree.md` — a row for the port and its fake, and for
  the five packages the table still does not name (`fence`, `hook`,
  `kitpaths`, `kitfetch`, `gitx`), which is `report-0003`.

## Outcome

_(fill after execution)_

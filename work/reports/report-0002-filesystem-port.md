---
id: report-0002
status: tracked
task_ref: [task-0016]
doc_ref: technical/engineering/boundaries.md
created: 2026-09-05T11:37:19Z
triaged: 2026-09-05T12:28:30Z
---

# The filesystem is the one boundary no command puts behind a port

**References:** [technical/engineering/boundaries.md](../../docs/technical/engineering/boundaries.md) · [task-0016](../tasks/task-0016-put-the-filesystem.md)

`docs/technical/engineering/boundaries.md` states that everything
leaving the process — script execution, `gh`, **the filesystem**, the
terminal — sits behind a small interface with a fake beside it. Three of
the four do: `gitRunner`, `Deps.Gh` and `Terminal`/`FakeTerminal`. The
filesystem does not: `initcmd`, `updatecmd` and `uninstallcmd` call
`os.ReadFile`, `os.WriteFile`, `os.MkdirAll`, `os.RemoveAll` and
`os.Stat` directly, about forty times.

The cost is measurable. Pushing coverage over `internal/` from 60.3% to
98.0% left 22 statements no fixture can reach, and 18 of them are error
returns from those direct calls: `filepath.Rel` and `DirEntry.Info`
inside a `WalkDir` callback, and the later writes of an `apply` whose
first write already failed — a sequence `chmod` cannot address one call
at a time. The tests that do reach the reachable ones work by making a
directory or a file read-only, which is a test shaping the machine
rather than the port it is testing.

Also observed: the function type `func(dir string, args ...string)
(string, error)` is declared three times — `kitfetch.GitRunner`,
`initcmd.gitRunner`, `hook.GitRunner` — so the wiring in
`cmd/writrun/main.go` converts between identical types.

---
id: report-0003
status: tracked
task_ref: [task-0016]
doc_ref: technical/layout/tree.md
created: 2026-09-05T10:48:47Z
triaged: 2026-09-05T12:50:58Z
---

# The layout table omits the packages this change extracted

**References:** [technical/layout/tree.md](../../docs/technical/layout/tree.md)

`docs/technical/layout/tree.md` lists one row per package under
`internal/`: `command`, `kit`, `forge`, `term`, `wrepo`. Implementing
task-0003 and task-0005 added five the table does not name — `fence`,
`hook`, `kitpaths`, `kitfetch` and `gitx` — and moved `ExecGit` and the
AGENTS.md graft out of `internal/command/initcmd/`. Both specs promise
`Proposed technical changes: none`, and
`writrun-check-spec-deltas spec-0003,spec-0005` exits 0 only because the
change leaves `docs/` alone; editing the table would make it exit 2,
UNDECLARED. Writing or changing anything under `docs/` is Thomas's gate
in `AGENTS.md`.

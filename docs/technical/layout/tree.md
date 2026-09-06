# Tree

| | |
|---|---|
| `cmd/writrun/` | Entry point — `main` and production wiring only. |
| `internal/command/` | One package per command: the use case, depending on ports and nothing else. |
| `internal/vfs/` | The filesystem — the port, the `os` implementation, and the fake. |
| `internal/kit/` | Runs the adopted repository's `.writrun/` scripts — the exec port. |
| `internal/gitx/` | One git invocation — the git port, and the type its consumers name. |
| `internal/forge/` | The `gh` invocations — the forge port. |
| `internal/term/` | Interaction: selection, confirmation, TTY detection — the terminal port. |
| `internal/wrepo/` | Adopted-repository detection, settings, queue paths. |
| `internal/fence/` | The fenced WritRun section of an `AGENTS.md`: grafted, replaced, removed. |
| `internal/hook/` | The commit-message hook: its text, where git keeps it, whether it is still ours. |
| `internal/kitpaths/` | What `init` installs, listed once — what `update` refreshes and `uninstall` removes. |
| `internal/kitfetch/` | The shallow clone of a pinned WritRun tag — the port, the clone, and the fake. |
| `internal/kittag/` | The tag `.writrun/VERSION` records — read once, ordered or matched. |
| `internal/requirements/` | What the wrapped scripts need on the PATH, listed once. |
| `scripts/` | This project's own machinery (the release path), not the methodology's. |
| `Makefile` | Thin aliases over `scripts/` and the suite — every target in [Makefile](makefile.md). |
| `tests/` | The suite — see [Testing](../testing/README.md). |
| `docs/` | The permanent documentation — [product](../../product/README.md) and technical. |

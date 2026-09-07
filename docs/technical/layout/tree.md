# Tree

| | |
|---|---|
| `cmd/writrun/` | Entry point — `main` and production wiring only. |
| `internal/command/` | One package per command: the use case, depending on ports and nothing else. |
| `internal/vfs/` | The filesystem — the port, the `os` implementation, and the fake. |
| `internal/kit/` | Names the adopted repository's `.writrun/` scripts and files, and runs the scripts — the exec port. |
| `internal/gitx/` | One git invocation — the git port, and the type its consumers name. |
| `internal/forge/` | The `gh` invocations — the forge port. |
| `internal/term/` | Interaction: selection, confirmation, TTY detection — the terminal port. |
| `internal/wrepo/` | Adopted-repository detection, settings, queue paths. |
| `internal/queue/` | The queue's folders, and its files as the kit's own scripts read them — front matter, one field, an inline list, the title, and the id-to-file resolution. |
| `internal/pointer/` | WritRun's section of an `AGENTS.md`: found, grafted, removed. |
| `internal/hook/` | The commit-message hook: its text, where git keeps it, whether it is still ours. |
| `internal/kitpaths/` | What `init` installs, listed once — what `update` refreshes and `uninstall` removes. |
| `internal/kitfetch/` | The shallow clone of a pinned WritRun tag — the port, the clone, and the fake. |
| `internal/kittag/` | Where `.writrun/VERSION` is, what it says, and how two tags order or match. |
| `internal/requirements/` | What the wrapped scripts need on the PATH, listed once. |
| `internal/chapter/` | Whether a docs folder holds a real chapter beyond its README. |
| `internal/screen/` | The queue as a key-navigated screen — what `writrun` with no command opens. |
| `scripts/` | This project's own machinery (the release path), not the methodology's. |
| `Makefile` | Thin aliases over `scripts/` and the suite — every target in [Makefile](makefile.md). |
| `tests/` | The suite — see [Testing](../testing/README.md). |
| `docs/` | The permanent documentation — [product](../../product/README.md) and technical. |

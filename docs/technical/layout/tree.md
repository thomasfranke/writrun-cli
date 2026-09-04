# Tree

| | |
|---|---|
| `cmd/writrun/` | Entry point — `main` and production wiring only. |
| `internal/command/` | One package per command: the use case, depending on ports and nothing else. |
| `internal/kit/` | Runs the adopted repository's `.writrun/` scripts — the exec port. |
| `internal/forge/` | The `gh` invocations — the forge port. |
| `internal/term/` | Interaction: selection, confirmation, TTY detection — the terminal port. |
| `internal/wrepo/` | Adopted-repository detection, settings, queue paths. |
| `scripts/` | This project's own machinery (the release path), not the methodology's. |
| `Makefile` | Thin aliases over `scripts/` and the suite; the scripts are the interface, CI calls them directly. |
| `tests/` | The suite — see [Testing](../testing/README.md). |
| `docs/` | The permanent documentation — [product](../../product/README.md) and technical. |

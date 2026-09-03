# Layout

| | |
|---|---|
| `cmd/writrun/` | Entry point — `main` and command wiring only. |
| `internal/` | Everything else. The public contract is the command line, never the Go packages. |
| `scripts/` | This project's own machinery (the release path), not the methodology's. |
| `tests/` | The suite — see [Testing](testing.md). |
| `docs/` | The permanent documentation — [product](../product/README.md) and technical. |

Nothing is exported for import: there is no `pkg/`, and `internal/`
enforces it. A consumer of this project runs the binary.

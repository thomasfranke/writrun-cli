# Public surface

Nothing is exported for import: there is no `pkg/`, and `internal/`
enforces it. The public contract is the command line, never the Go
packages; a consumer of this project runs the binary. Each port's
interface is defined in the package that consumes it, with a fake
beside it — [boundaries](../engineering/boundaries.md).

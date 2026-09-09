# Public surface

Nothing is exported for import: there is no `pkg/`, and `internal/`
enforces it. The public contract is the command line, never the Go
packages; a consumer of this project runs the binary. Each port's
interface is defined in the package that consumes it, with a fake
beside it — [boundaries](../engineering/boundaries.md).

One package is shared rather than consumed through a port:
`internal/palette` decides what a colour means, and every command that
paints anything asks it. It is not a port — there is nothing to fake and
no boundary to cross — but it is the same rule one level down: a
vocabulary stated once, so two commands cannot drift into two of them.
Its zero value paints nothing, so a caller never branches on whether
colour is on ([rules](../../product/rules.md)).

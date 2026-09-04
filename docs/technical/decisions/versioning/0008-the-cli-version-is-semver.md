# the CLI's version is SemVer; WritRun's scheme stays upstream.

**2026-09-03**

`v0.0.01` is not valid SemVer (leading zero): the Go module system
ignores the tag, `go install @latest` resolves to a pseudo-version of
`HEAD`, and GoReleaser refuses the cut. Rejected: sharing WritRun's
scheme (the symmetry is aesthetic — the binary shows its own version
and the pinned tag side by side either way) and dropping the
`go install` route (loses a whole install path to keep a digit).

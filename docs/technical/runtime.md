# Runtime

- **Go**, one static binary per platform. Binary: `writrun`. Module
  path: `github.com/thomasfranke/writrun-cli`.
- Standard library first — no CLI framework until a command proves it
  necessary.
- `writrun` adds no runtime requirement. The wrapped scripts keep
  theirs: `git`, `bash`, POSIX `awk`/`sed`, and `gh` where the flows
  already use it.
- Supported platforms: macOS, Linux, Windows — whatever the Go
  toolchain cross-compiles without cgo.

# Distribution

- Homebrew: `brew install thomasfranke/tap/writrun` — the formula is
  `writrun`, the installed binary is `writrun`.
- Prebuilt binaries per platform on tagged GitHub releases.
- `go install github.com/thomasfranke/writrun-cli/cmd/writrun@latest`.
- GoReleaser cuts releases from tags on `main` and updates the tap
  formula; each release names the WritRun tag it targets.

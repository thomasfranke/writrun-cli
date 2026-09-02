# Decisions

Dated, append-only, one global log for the whole project. Never edit an
entry; add the next one.

- **2026-08-24 — Go, single binary, standard library first.** Rejected:
  Rust (equivalent fit, higher cost for a thin wrapper), Bash (no
  build, but flags/help/distribution become artisanal), interpreted
  languages (a runtime requirement the methodology never had).
- **2026-08-24 — the operated repository's scripts are the execution
  authority.** Rejected: embedding them in the binary — a second source
  of truth that drifts from what CI runs and overrides the tag the
  repository pinned.
- **2026-08-24 — agent command via `git config writrun.agent`.**
  Rejected: a file in `.writrun/` (imposes one agent on every
  contributor, mixes client config into the methodology's home) and a
  CLI-owned config file (a new file for one key git config already
  layers correctly).
- **2026-08-24 — kit fetched by shallow `git clone` at the pinned
  tag.** Rejected: GitHub tarball download (adds an HTTP path when git
  is already required) and `go:embed` (same second-source-of-truth
  problem as embedding scripts).
- **2026-08-24 — the binary is `writrun`.** One name across project,
  formula, and binary. Rejected: `writ`, the short form WritRun's About
  suggests — two names for one tool cost more than the four letters
  saved.
- **2026-08-24 — one technical file per subject, README as index.**
  Separate files stay readable and are maintained independently; a
  single growing file makes every `doc_ref` point at the same place.

# Architecture

- Every check or algorithm runs as a child process: the **operated
  repository's own `.writrun/` scripts**. Nothing is reimplemented in
  Go, nothing embedded in the binary. If binary and scripts disagree,
  the scripts are right.
- `writrun` builds on WritRun's declared public contract (its
  `docs/technical/README.md`, Distribution): front-matter schemas, the
  `docs/` + `work/` split, script arguments and exit codes, the
  grep-level markers.
- The contract is alpha. Each `writrun` release **pins one WritRun
  tag**; `.writrun/VERSION` in the operated repository says which tag
  its copy came from. A mismatch is reported, never silently bridged.
- `writrun work` reads the adopter's agent command from
  `git config writrun.agent` — repo and user layers for free, nothing
  committed by accident.
- `writrun init` / `writrun update` obtain the adoption kit by
  `git clone --depth 1 --branch <tag>` of the WritRun repository and
  copy `template/` from it.

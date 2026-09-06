# Decisions

**The dated why behind each piece of machinery — and what was
rejected.** Append-only history: an entry is never edited, the next one
is added. The living reference is the rest of
[`technical/`](../README.md); this folder is where its shape came from.

One decision per file, numbered in the order it was taken. **The number
is identity**: it never changes, the file is never renamed, and a
superseded decision keeps its file and its number — the entry that
replaces it says so and gets the next number.

Entries sit in the folder of the technical subject they concern —
the taxonomy the [`technical/` index](../README.md) already fixes.
**The table below is the chronology**, which folders do not carry; it
is the only part of this folder that is rewritten, and only by
appending a row.

| # | Date | Subject | Decision |
|---|---|---|---|
| [0001](runtime/0001-go-single-binary.md) | 2026-08-24 | runtime | Go, single binary, standard library first. |
| [0002](architecture/0002-scripts-are-the-authority.md) | 2026-08-24 | architecture | the operated repository's scripts are the execution authority. |
| [0003](architecture/0003-agent-command-via-git-config.md) | 2026-08-24 | architecture | agent command via `git config writrun.agent`. |
| [0004](architecture/0004-kit-by-shallow-clone.md) | 2026-08-24 | architecture | kit fetched by shallow `git clone` at the pinned tag. |
| [0005](runtime/0005-the-binary-is-writrun.md) | 2026-08-24 | runtime | the binary is `writrun`. |
| [0006](docs/0006-one-file-per-subject.md) | 2026-08-24 | docs | one technical file per subject, README as index. |
| [0007](engineering/0007-google-style-ports-three-tiers.md) | 2026-09-03 | engineering | Google Go style, ports at the boundaries, three test tiers gated in CI. |
| [0008](versioning/0008-the-cli-version-is-semver.md) | 2026-09-03 | versioning | the CLI's version is SemVer; WritRun's scheme stays upstream. |
| [0009](runtime/0009-interaction-via-charm-huh.md) | 2026-09-03 | runtime | terminal interaction via the Charm stack, `huh` at the surface. |
| [0010](docs/0010-help-is-one-line-per-command.md) | 2026-09-03 | docs | `--help` is one line per command plus the docs' address. |
| [0011](docs/0011-one-decision-per-file.md) | 2026-09-03 | docs | one decision per file, numbered, in the subject's folder. |
| [0012](docs/0012-technical-subjects-are-folders.md) | 2026-09-04 | docs | technical subjects are folders; `architecture.md` alone stays at the root. |
| [0013](architecture/0013-the-kit-is-read-from-the-kit.md) | 2026-09-06 | architecture | the kit is read from the kit; Go names only what the binary calls. |

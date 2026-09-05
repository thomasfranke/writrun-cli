---
id: report-0012
status: open
task_ref: []
doc_ref: null
created: 2026-09-05T18:21:19Z
triaged: null
---

# The stage-0 and stage-1 checks are written twice, in initcmd and in doctorcmd

`internal/command/initcmd/checks.go` and
`internal/command/doctorcmd/files.go` each carry their own copy of the
stage-0 PATH probe, the About-file, product-chapter, technical-doc and
`work/` folder checks, the `AGENTS.md` marker check and the
`.writrun/VERSION` check. `initcmd/checks.go` calls its output a gap and
never blocks adoption; `doctorcmd` calls it a finding, grades it, and
sets the exit status from it. The two disagree today on three points:
doctor reads the declared stage through the repository's own
`read_setting.sh` where init takes it from the flag it just asked for,
doctor runs `check_front_matter.sh` and `check_settings.sh` where init
runs neither, and doctor names the four human gates one by one where
init tests `AGENTS.md` for the string `<!-- TODO`. The duplication was
left standing because four tasks were in flight on the same packages.

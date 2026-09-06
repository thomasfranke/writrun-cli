---
id: report-0024
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-06T01:30:28Z
triaged: 2026-09-06T02:23:26Z
---

# The Go front-matter reader accepts files the kit's own reader calls malformed

`frontMatter` in `internal/command/amendcmd/queue.go` (and its identical
twin in `finishcmd`) opens the block on `strings.TrimSpace(lines[0]) ==
"---"`, while the kit's `ql_fm_field` in
`.writrun/scripts/stage-2-pull-requests/queue_lib.sh` opens it on
`NR == 1 { if ($0 != "---") exit }` — no trimming. A CRLF queue file
therefore has front matter to the Go reader and none to the shell one:
running `check_front_matter.sh` over such a file reports "MALFORMED:
front matter must open at line 1 with --- and close with ---", and
`writrun amend` reads its `status`, rewrites it and opens the pull
request over the same bytes. The comment above `field` says it keeps
"the same rule ql_fm_field keeps", which is true of where the block ends
and not of where it begins. Changing the Go side alone would leave two
copies of the helper disagreeing, which is why it was left as observed.

**Triage — fixed.** Both Go readers now match the fence exactly, the
way `ql_fm_field` does: `frontMatter` in `internal/command/amendcmd/`
and in `internal/command/finishcmd/` open the block only on a line that
is `---` and nothing else. A CRLF queue file therefore has no front
matter here either, so `amend` refuses it where it used to read a
status off it and write one back. The objection that stopped the fix
when this was written — that changing one copy would leave the two
disagreeing — was answered by changing both; they were byte-identical,
and the duplication itself stays with report-0020.

`setField`'s line-ending preservation, added in the same change, keeps
its point for a fence that is the kit's and a field line that still
carries a carriage return. Its test moved to that case, because the
file it used to be given is one the repository calls malformed and this
command now refuses.

---
id: report-0024
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T01:30:28Z
triaged: null
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

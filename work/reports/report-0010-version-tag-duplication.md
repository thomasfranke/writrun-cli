---
id: report-0010
status: tracked
task_ref: [task-0019]
doc_ref: null
created: 2026-09-05T18:20:11Z
triaged: 2026-09-05T19:11:12Z
---

# Three packages will read .writrun/VERSION and compare tags

**References:** [task-0019](../tasks/task-0019-consolidate-duplication.md)

`updatecmd` reads `.writrun/VERSION` in `recordedTag` and orders two
tags in `compareTags`/`parseTag`; `statuscmd`, added by task-0014, now
reads the same file and compares the same two values with its own
`recordedTag`, `sameRelease` and `numbers` — it needs equality, not
order, so the copy is smaller rather than identical. task-0004
(`doctor`) is specified to read that file too and would make a third.
The duplication was taken deliberately: the three commands are being
written in parallel, and a shared package each would have to agree on
before any could land is a coupling none of them can pay for yet. What
is written down here is that after all three have landed, one file has
three readers and one comparison has two implementations that already
disagree on what they compute.

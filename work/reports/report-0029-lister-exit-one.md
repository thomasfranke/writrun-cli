---
id: report-0029
status: open
task_ref: []
doc_ref: null
created: 2026-09-08T15:06:47Z
triaged: null
---

# The screen read the lister's exit 1 as a failure, not as its answer

`make ui` failed in this repository with `writrun: exit status 1`, and
the screen never opened. Reproduced under a pseudo-terminal on
2026-09-08 at 8470a17.

The lister's last statement is `[ -n "$available" ]`, so its exit status
answers "is anything available" — 1 is the answer "no". Three callers
read that same script and only one read the status as a failure:

- `internal/command/listcmd/listcmd.go:80` maps exit 1 to nil, with a
  comment saying nothing available is an answer to what was asked.
- `internal/command/takecmd/takecmd.go:138` guards on
  `exitCode(err) != 1`, with a comment saying the same.
- `cmd/writrun/main.go:175` returned the error, and
  `internal/command/run.go:238` printed it.

So the screen opened only where a task was available. That the opposite
was intended is visible in `internal/screen/model.go:146`, whose
`nothing to select · s status · q quit` footer is written for the empty
queue and was unreachable.

Fixed in the change that carries this report, by the guard the other two
callers already use. The finding outlives the fix: one script's exit
code is read in three places, and nothing makes the three agree.

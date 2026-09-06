---
id: task-0021
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0020]
doc_ref: product/screen.md
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-09-06T02:52:31Z
queued: 2026-09-06T03:00:56Z
completed: null
merged: null
provenance: []
---

# Navigate the queue with writrun and no command

**References:** [product/screen.md](../../docs/product/screen.md) · [spec-0020](../specs/spec-0020-queue-screen.md)

`writrun` with no command prints what `--help` prints. `product/screen.md`
says it should open the queue as a screen navigated by keys, and reserves
the help text for the two cases where a screen cannot exist: no terminal,
or no adopted repository. Today those two exceptions are the only
behaviour there is.

The rule was authored and merged in its own pull request and has stood
since without deriving anything, so the gap is not a half-finished
implementation — it is a rule waiting for the work it authorizes. Every
other line of the documented product has code behind it; this is the one
that does not.

What it buys is the entry point the rest of the product already assumes.
`list` shows the queue and stops; taking the task it just showed means
reading an id off the screen and typing it back. The screen closes that
loop without inventing anything: it offers no action a command does not
already provide, and every change still goes through the command it
dispatches, with that command's own checks, questions and confirmation.

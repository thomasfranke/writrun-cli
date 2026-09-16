---
id: task-0036
status: in-review
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0044]
doc_ref: product/screens/README.md
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-16T19:12:48Z
queued: 2026-09-16T23:03:38Z
completed: 2026-09-16T23:35:41Z
merged: null
provenance: []
---

# Give a command that asks its own process, so no reader outlives it

**References:** [product/screens/README.md](../../docs/product/screens/README.md) · [spec-0044](../specs/spec-0044-command-own-process.md)

A command chosen on the screen that asks a question runs on the terminal
the screen released, and when the screen comes back the next key the
reader presses can do nothing at all. The question's input reader
survives the question, takes that key, and drops it — it is handing
messages to a program that has ended.
[report-0040](../reports/report-0040-swallowed-key.md) carries the
evidence and names why nothing pinned here can be moved to escape it.

Make the command a process of its own. `screens/README.md` already
states the rule the defect breaks — "a command owns the keyboard alone",
and "two of those in one process do not share a keyboard" — and answers
it by pausing the screen and releasing the terminal. That is the half
that is not enough: releasing the terminal does not end the reader that
was holding it, and the two programs go on sharing one keyboard for as
long as that reader takes to die. A process ends its own readers when it
exits, which is the only thing that makes the rule true rather than
intended.

It matters beyond the one lost key. The reader who presses `q` and is
answered with nothing has no way to tell a swallowed key from a hung
screen, and the honest guess is the wrong one — they press it again and
it works, so the screen looks unreliable rather than busy. And it is not
only a reader's problem: the same defect has taken this repository's
pipeline red on four pull requests, three of them carrying no code at
all, so every change now pays a re-run to find out whether it was the
change or the dice.

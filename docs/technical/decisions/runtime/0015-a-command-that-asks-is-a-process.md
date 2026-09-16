# a command that asks runs as its own process.

**2026-09-16**

A command chosen on a screen that asks a question is spawned — this
binary, run again on the terminal this one holds — rather than called
inside the process the screen is running in. The screen's hand-over is
unchanged: it still pauses, releases the terminal, names what it is
running and waits for the return.

The reason is that releasing a terminal does not end the reader holding
it. A terminal program cancelled while already waiting on the terminal
can go on holding a blocking read, take the next key, and deliver it to
a program that has ended, where it is dropped without a word — so the
first key after a question does nothing
(`work/reports/report-0040-swallowed-key.md`; the race is
[bubbletea#1116](https://github.com/charmbracelet/bubbletea/issues/1116),
open, in a package already pinned at its newest release). A process ends
its own readers when it exits, and nothing available here does so
reliably otherwise.

Rejected: waiting for the library, which would leave the defect in the
hands of a repository we do not own; pinning a patched fork of
`cancelreader`, which buys one fix and a dependency nobody reviews;
giving the question a file of its own to read, which fails because a
second handle on the same terminal reads the same keys; and asking the
reader to press the key twice, which is not a decision but a symptom
written down.

The costs are named rather than denied. A question now costs a process,
which is a few milliseconds nobody waiting to type will notice. The
frame's flags have to travel to the child, or a session running under
`--yes` would ask what its parent had answered. And a binary that cannot
name itself has no child to spawn, so the command runs in this process
as it always did — the old risk, kept deliberately, because a screen
that refuses to open a question is worse than one that may swallow a key.

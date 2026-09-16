---
id: report-0040
status: tracked
task_ref: [task-0036]
doc_ref: product/screens/README.md
created: 2026-09-16T19:12:39Z
triaged: 2026-09-16T19:30:00Z
---

# the first key after a question is sometimes swallowed

**References:** [product/screens/README.md](../../docs/product/screens/README.md)

A command chosen on the screen that asks a question — `report`, or a
change on the config screen — runs on the terminal the screen released.
When it is done and the screen comes back, the next key the reader
presses does nothing. Pressing it again works.

It is roughly one run in ten on Linux, and never once on macOS. It has
taken this repository's own pipeline red on four pull requests —
[#129](https://github.com/thomasfranke/writrun-cli/pull/129) needed
three attempts, [#139](https://github.com/thomasfranke/writrun-cli/pull/139)
and [#141](https://github.com/thomasfranke/writrun-cli/pull/141) two
each — at two different e2e cases, always with the same last words:
`q did not leave the screen`.

The screen is not hung. A build logging every message the session
receives shows it still working after the command returned — a
`WindowSizeMsg`, a `tickMsg` — and no `KeyMsg` for the key that was
pressed. A goroutine dump taken eight seconds after the keypress shows
one input reader in the process, blocked inside `read(2)` on the
terminal: it had been told a byte was waiting, and by the time it read,
the byte was gone.

A build with the input library instrumented names what took it:

    CR#2 wait -> <nil>              the question's own reader
    CR#2 read n=1 data="\x1b"       reads the esc that cancels it
    CR#2 cancel                     the question ends, its reader is cancelled
    CR#2 wait -> <nil>              but the wait answers "readable", not "cancelled"
         waitForLine read "\n"      and the screen's own read takes the return
    CR#3 created fd=0               the screen comes back, with a new reader
    CR#3 wait -> <nil>              which is told a byte is waiting
    CR#2 read n=1 data="q"          and the cancelled reader is the one that gets it

The question's reader is cancelled while already inside its wait, and
that wait answers "there is a byte" rather than "you were cancelled".
It enters a blocking read — and the byte it was woken for is the return
the reader pressed to come back, which `internal/screen/session.go`'s
`waitForLine` reads itself, on the same file descriptor, to know when to
redraw. So the cancelled reader sits in a read with nothing to read, and
takes the next key instead. It hands that key to a program that has
already ended, where the message is dropped without a word. The screen's
own reader, woken by the same byte, finds the terminal empty and blocks
— which is the state the goroutine dump caught.

Two things are needed, and this path always has both: a question that is
cancelled, and a read of the screen's own on the same descriptor
immediately after. That is why it is always the first key after the
question and never the return that precedes it.

The cancellation race is the input library's —
[charmbracelet/bubbletea#1116](https://github.com/charmbracelet/bubbletea/issues/1116),
open and unfixed, with `muesli/cancelreader` v0.2.2 the newest release
of the package that loses the byte. Nothing pinned here can be moved to
escape it. What is this repository's is the second reader:
`session.go` says a command that asks is handed the terminal "because
two programs must never hold one keyboard", and then keeps a reader of
its own in the same process while the question's reader is dying.

**Triage — tracked.** [task-0036](../tasks/task-0036-command-own-process.md)
carries it, through [spec-0044](../specs/spec-0044-command-own-process.md):
a command that asks runs as its own process, so no reader of its can
outlive it. Measured against the same instrumented build, which widens
the race until it fires about half the time: as it stands, 5 failures in
11 runs; with the command spawned as a process, 0 in 20.

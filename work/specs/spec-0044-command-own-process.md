---
id: spec-0044
task_ref: task-0036
status: approved
created: 2026-09-16T19:12:57Z
---

# spec-0044 — A command that asks runs as its own process

**References:** [task-0036](../tasks/task-0036-command-own-process.md)

- **Goal:** a command that asks is a process, so no reader of its can outlive it.

## Scope

In: how a command that asks is run when a key on a screen chooses it,
and the port that running goes through.

Out: the screens themselves — their keys, their drawing and their
hand-over are unchanged. Out: the input library's cancellation race,
which is upstream's, open, and not fixed here — this spec makes it
unreachable rather than correct. Out: `writrun <command>` typed at a
shell, which is already a process of its own and never had the defect.

## Steps

1. A command that asks runs as a process of this binary rather than a
   call inside it. Everything the reader sees is unchanged: the screen
   still pauses, still releases the terminal, still says what it is
   running, and still waits for the return before it redraws.
2. The spawning is a port — defined where it is consumed, with a fake
   beside it, and wired in `cmd/writrun/` only, as
   `technical/engineering/boundaries.md` requires of everything that
   leaves the process. The fake is what keeps every existing case that
   drives a screen driving the dispatch it was written against.
3. The child is given what the parent was given: the command, the one
   argument a row may carry, and the flags the session is running under.
   A question answered in the child must be the question the parent
   would have asked.
4. Where the running binary cannot be named, the command runs in this
   process as it does today. A screen that cannot open a question is
   worse than one that may swallow a key.
5. A command that asks nothing is untouched. It is captured in this
   process, it opens no terminal program of its own, and it has no
   reader that could outlive it.

## Acceptance criteria (EARS)

- When a key on a screen chooses a command that asks, the system shall
  run that command as a process of its own.
- When that process has exited and the reader presses the return, the
  system shall redraw the screen it was chosen from, as it does today.
- When a key is pressed on the redrawn screen, the system shall act on
  that key — the first one, not the second.
- When a key on a screen chooses a command that asks nothing, the system
  shall capture it in this process, unchanged.
- When the running binary cannot be named, the system shall run the
  command in this process rather than refuse it.

## Edge cases

- **The binary is replaced or removed while the screen is open.** The
  name answers a path that no longer runs, and spawning fails. It is the
  same answer as not being able to name it: the command runs in this
  process, and the reader is not told about a failure that changed
  nothing they can see.
- **The command refuses.** Its exit code is the child's, and the screen
  already ignores a command's exit code — a refusal inside a session is
  read on the terminal, not by a script. Nothing here changes that.
- **The suite's own terminal.** A case driving a form through
  `WRITRUN_TTY_IN` keeps working: the child inherits the environment,
  so the bytes the case wrote are the bytes the form reads.

## Tests required

Unit, `internal/command`: that a command declaring it asks goes through
the port, that a command declaring it asks nothing does not, and that a
binary which cannot name itself falls back to running in this process.

End-to-end, `tests/e2e/screen/`: the two cases that already drive a
question and then press a key —
`the_screen_answers_keys_test.sh` and `the_config_screen_test.sh` — are
the assertion, and they do not change. What changes is that they stop
being a coin toss, and that is measured rather than assumed: see the
Definition of Done.

## Definition of Done

- [ ] A command that asks is run through the port, and no package under
      `internal/screen/` or `internal/command/` names an executable.
- [ ] The two e2e cases above pass twenty consecutive runs on Linux.
      The same measurement against the binary as it stands is what
      [report-0040](../reports/report-0040-swallowed-key.md) records, so
      the two numbers are comparable.
- [ ] [report-0040](../reports/report-0040-swallowed-key.md) closed.

## Proposed product changes

- `product/screens/README.md#writrun-with-no-command` — the rule that a
  command owns the keyboard alone says how it is kept. Pausing the screen
  and releasing the terminal is the half that is not enough: releasing
  does not end the reader that was holding it.

## Proposed technical changes

- `technical/engineering/boundaries.md` — the list of what leaves the
  process gains the binary itself.
- `technical/decisions/runtime/0015-a-command-that-asks-is-a-process.md`
  — the decision, what it costs, and what it does not fix.
- `technical/decisions/README.md` — the index gains the entry, because a
  decision nobody can find from the list is one the next reader derives
  again.

## Outcome

_(fill after execution)_

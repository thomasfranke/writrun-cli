---
id: spec-0021
task_ref: task-0022
status: implemented
created: 2026-09-06T02:53:08Z
---

# spec-0021 — Surviving a signal mid-finish

**References:** [task-0022](../tasks/task-0022-signal-window.md)

- **Goal:** a catchable signal arriving between `finish`'s completion writes and its confirmation runs the same undo every other non-success end runs, and the process still dies of that signal.

## Scope

In: `internal/command/finishcmd` — the window from the completion writes
to the end of the confirmation, and the journal that already knows how
to put those writes back. Out: the order of the writes, which spec-0017
settled and this does not reopen; every other command; the undo's
behaviour on the ends it already reaches, which is unchanged.

## What is wrong

The undo added by spec-0017 is ordinary control flow — a deferred call
on paths the function returns through. A signal returns through none of
them. report-0018 reproduced it: SIGTERM two seconds into a slowed
`preflight.sh`, exit 143, `git status --porcelain` showing both queue
files changed, the spec `implemented` and the task carrying its
completion date.

That is exactly the state `product/pull-requests/shape.md` forbids of a
refused command and exactly the finding report-0015 raised, standing
again on the one path spec-0017's undo does not cover. spec-0017 named
the dirty window as the cost of keeping the spec's order; it did not
say the window was un-guarded, and a reader of its Outcome would not
learn that a Ctrl-C leaves the tree changed.

Ctrl-C at the confirmation is not this bug and must not be broken by
fixing it: huh reads the interrupt as a key, returns an error, and the
undo runs today.

## Steps

1. Arm a handler for the catchable termination signals — at minimum SIGINT and SIGTERM — when the journal takes its first entry, and disarm it when the command reaches an end on its own.
2. On a signal, run the same `journal.restore` every other non-success end runs, including its check that a file still holds what this run left, and its message when it does not.
3. Die of the signal after restoring: reset the handler to the default and re-raise, so the exit status is the shell's conventional 128+n rather than a code invented here.
4. Report a restore that fails on this path the way the other paths report it — naming the files left changed and telling the user to put them back by hand — before the process dies.
5. Leave the confirmation's own interrupt handling alone: huh keeps reading Ctrl-C as a key while it holds the terminal, and that path already restores.

## Acceptance criteria (EARS)

- When SIGTERM arrives between the completion writes and the end of the confirmation, the system shall restore both files before exiting.
- When SIGINT arrives in that window and no prompt is reading the terminal, the system shall restore both files before exiting.
- When the system exits because of a signal, it shall exit with the status that signal conventionally produces, not with a code of its own.
- When a signal arrives before the first completion write, the system shall exit without restoring anything, there being nothing to put back.
- When a signal arrives after the pull request is opened, the system shall not restore the completion writes.
- When the restore fails on the signal path, the system shall name the files left changed before exiting.
- When no signal arrives, every existing end shall behave exactly as it does today.

## Edge cases

- SIGKILL and a lost machine cannot be caught, so the window narrows rather than closing; the residue is the same two files and `writrun status` is what finds it.
- A second signal while the restore is running: the restore is short and must not be restarted by it — the first one wins and the second does not leave the tree half put back.
- The signal arrives while huh holds the terminal: the prompt's own interrupt path already restores, and the handler must not double-restore.
- A signal during `preflight.sh`: the child gets the signal too, and its exit is not the answer — the undo runs regardless of what the script reports on its way out.

## Tests required

Integration driving a real `writrun finish` against the fixture with a
slowed `preflight.sh`, sending SIGTERM into the window and asserting
`git status --porcelain` is empty afterwards, the spec still `approved`
and the task's `completed` still null — the assertions report-0018's
reproduction failed. One case for SIGINT on the same path. One asserting
the exit status is the signal's. One asserting a signal before the first
write restores nothing and reports nothing. The existing `finish` suite
passes untouched, which is what proves the handler changed no other end.

## Definition of Done

- [x] report-0018's reproduction, run again, leaves a clean working tree.
- [x] The process still dies of the signal, with the signal's own status.
- [x] Every existing `tests/integration/finish/` case passes unchanged.

## Proposed product changes

- none — `shape.md` already states the rule this makes true on one more path.

## Proposed technical changes

- none — no new package and no new boundary; the journal and the `vfs` port already exist.

## Outcome

**The guard is a goroutine, and the journal is what makes it safe.**
`finishcmd/signals.go` arms a handler the moment the journal holds what
it would put back — after the last `remember`, before the first write —
and `run` disarms it on every end it reaches on its own. A caught signal
runs the same `journal.restore` every other non-success end runs, writes
the same sentence about anything it could not put back, and then dies of
the signal. `Die` resets the disposition and raises the signal again, so
the status is 130 or 143 rather than a code invented here.

**Two signals, and the set is the decision.** SIGINT and SIGTERM,
nothing else. A signal this command takes off its default disposition is
one it must then answer for, and those two are the ones huh already
answers while it holds the terminal, which is what makes standing down
for the question free. SIGHUP was in the set for a while and came out:
nothing else answers it, so the guard could not stand down for it at the
question, and a hangup the guard then dropped would have left the
command waiting at a prompt on a dead terminal — a hang where there had
been a death. SIGQUIT keeps its default, which is the runtime's
goroutine dump.

**Standing down for the question is `signal.Stop`, not a flag.** The
first attempt was a flag the watcher read after taking the signal off
the channel, and it lost the race under `-race` on every run: the
question returns, the flag flips, and the watcher then answers a signal
that was the prompt's. `signal.Stop` is the only thing that guarantees
the channel receives nothing more. A second registration holds the
disposition for the length of the question, so standing down never hands
a signal back to the default and kills the process outright.

**One restore is the journal's guarantee, not the stand-down's.** A
signal the prompt answered can still be handed over after the question
returns — `signal.Stop` unregisters, but the runtime processes what it
has already queued whenever it gets to it — and a signal during
`preflight.sh` reaches the child too, so the script's non-zero verdict
asks for the same undo on the way up. `journal.restore` now runs at most
once under a mutex: the second caller waits for the first and finds the
work done, so no restore is left half made and nothing is put back
twice. The same mutex guards `remember`, `left` and `unknown`, which the
command's goroutine writes while the guard's reads.

**A signal after the pull request is ready takes nothing back.**
`journal.seal` closes the undo the statement after `markReady` returns
nil. Without it the window between the forge answering and `run`
returning was a window in which the guard would have reverted a finish
that had succeeded.

**A signal this process ignores is not armed for.** `armable` filters
the set through `signal.Ignored`, and that is not a nicety: a shell
starts a background job with SIGINT ignored, so `writrun finish &` in a
script inherits it. Arming there would have put the completion writes
back under a run that then went on to succeed, and the re-raise would
have found the same ignore and killed nothing — the run hung instead of
dying. The suite reproduced it before the filter existed, which is why
`finish_into_a_signal` runs under `set -m`.

**Where the status is not the signal's, and why that stands.** Criterion
3 asks for 128+n whenever the system exits because of a signal, and step
5 says to leave the confirmation's own interrupt handling alone. At the
question the two point different ways: bubbletea answers the signal, huh
returns, and the command ends at the decline's exit 1 with the tree put
back. Step 5 is the explicit instruction, so it wins, and the outcome is
recorded here rather than papered over. Every other point in the window
exits at the signal's own status.

**What is left open.** SIGKILL cannot be caught and a lost machine
cannot be answered, so the window narrows rather than closing; the
residue is the same two files and `writrun status` is what finds it.
Two instants stay uncovered inside it: between `gh pr ready` answering
and `seal`, and between `signal.Stop` at the question and bubbletea's
own registration.

Tests: `signals_test.go` drives a real signal into a real `run` — the
fake `preflight.sh` raises it and waits for the death — and covers both
signals, the restore reported once when the script also fails, the undo
running whatever the script reports on its way out, the failed restore
naming the file left changed on stderr, a signal before the first write
saying nothing, a sealed journal taking nothing back, the ignored
signal, the stand-down at the question and the delivery coming back
after it, and a disarmed guard leaving a signal alone.
`tests/integration/finish/` gains report-0018's reproduction on the real
binary against the fixture with a slowed `preflight.sh`: the terminate
at 143, the interrupt at 130, `git status --porcelain` empty, the spec
still `approved`, the task's `completed` still null — and a third case
sending the signal into a slowed `check_deltas.sh`, before the first
write, asserting nothing is put back and nothing is said. Both `Deps`
gain `Die`, so no case kills the test binary. The fixture's captured
streams and fake tree took the fixture's own lock, because production
answers two goroutines with an `*os.File` and the disk and the fakes
standing in for them have to answer the same way.

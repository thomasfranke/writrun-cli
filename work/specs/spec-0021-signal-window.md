---
id: spec-0021
task_ref: task-0022
status: draft
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

- [ ] report-0018's reproduction, run again, leaves a clean working tree.
- [ ] The process still dies of the signal, with the signal's own status.
- [ ] Every existing `tests/integration/finish/` case passes unchanged.

## Proposed product changes

- none — `shape.md` already states the rule this makes true on one more path.

## Proposed technical changes

- none — no new package and no new boundary; the journal and the `vfs` port already exist.

## Outcome

_(fill after execution)_

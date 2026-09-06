---
id: spec-0017
task_ref: task-0018
status: implemented
created: 2026-09-05T19:11:32Z
---

# spec-0017 — Settle when finish writes

**References:** [task-0018](../tasks/task-0018-finish-write-order.md)

- **Goal:** `writrun finish` leaves nothing behind when it is refused, and the checks it runs judge the edits it made.

## Scope

In: the order of spec-0010's five steps — when the two completion
writes happen relative to the confirmation and to `preflight.sh`; what
`finish` does with the edits it makes. Out: what the checks themselves
assert (`preflight.sh` and the scripts under it are the kit's, and this
repository does not change them); the merge; the task's `status` line.

## The two statements in conflict

`shape.md` says a refused pull-request command leaves nothing behind —
"no half-written status, no orphan branch". spec-0010 fixes an order in
which the writes are step 2 and the confirmation is step 5, because
step 4's `preflight.sh` reads the `completed` date the writes leave.
Both cannot hold. The decision this spec records is which one gives.

## Steps

1. Decide the order, and say why in the Outcome. Two candidates, and the choice is the work:
   - **Confirm first.** The question moves ahead of the writes; `preflight.sh` then runs against a tree that already carries them, and a decline never writes. The cost is that the composition shown at the prompt cannot quote a check that has not run.
   - **Write, check, then undo on a decline.** The order stands and a refusal restores the two files. The cost is a restore path that has to be right, and a window in which the tree is dirty.
2. Apply the decision to `finishcmd`, and to spec-0010's Steps — an approved spec that was out-implemented is stale, so it is amended through `draft` rather than quietly diverged from.
3. Make the completion edits reach the range the later checks read, or state plainly in the Outcome that they cannot and what the gates therefore vouch for.
4. Give `a_decline_reaches_nothing_test.sh` the whole sentence to assert: the forge untouched **and** the queue files unchanged.

## Acceptance criteria (EARS)

- When the user declines, the spec's `status` and the task's `completed` date shall be exactly what they were before the command ran.
- When `finish` completes, the checks it ran shall have read the completion edits, or the Outcome shall name what they read instead.
- When the command is rerun after a decline, it shall behave as it did the first time.
- When the order changes, spec-0010's Steps shall change with it.

## Edge cases

- A decline after a partial write: whichever order is chosen, the tree must end where it started, and a failure to restore must say so rather than exit quietly.
- `--yes`: no prompt, so the decline path is unreachable — the order must still leave the checks reading the edits.
- A task with no spec: there is one write, not two, and the same guarantee holds.

## Tests required

Unit: the decline path asserted over both files, not just the forge.
Integration: a declined `finish` leaving `git status --porcelain` empty;
a confirmed `finish` whose checks demonstrably saw the edits.

## Definition of Done

- [x] A declined `finish` leaves the working tree as it found it, proven by a test.
- [x] The order is recorded in the Outcome with the reason it was chosen.
- [x] spec-0010's Steps agree with what `finishcmd` does.
- [x] Suite green.

## Proposed product changes

- none — `shape.md` already states the rule; this change makes the
  implementation obey it.

## Proposed technical changes

- none.

## Outcome

**The order stands; the writes are undone.** `finishcmd` keeps
spec-0010's five steps exactly where they were — deltas, the two
completion writes, the ledger, `preflight.sh`, then the composition and
the question — and every end after step 2 that is not a success puts the
two files back byte for byte before it exits: a refused ledger, a
non-zero preflight, a forge that will not answer, a pull request already
merged, and the no at the question. spec-0010 gains a step 6 saying so.

Three reasons the other candidate lost. **`shape.md` fixes this order in
its own sentence** — "run the repository's checks …, write the status the
flow calls for, assemble …, show all of it, and open the pull request on
confirmation" — so confirming first would have taken a product-doc
change, and this spec promises none. **Confirming first only moves the
leftovers.** The writes would still have to precede `preflight.sh`, so a
gate failing after a yes leaves exactly the two edits report-0015 found;
the guarantee would hold for a decline and break for a red gate, while
the undo holds for both. And **it would ask before the gates had
spoken** — the composition could not quote preflight's verdict, which is
most of what the question is for.

A restore that fails is the one case that rewrites the verdict: it
returns an error naming the files left changed and telling the user to
put them back by hand, rather than the decline or the script's exit code,
because the frame passes an exit code up without printing a word — and
"declined — nothing changed" over a tree that did change is the sentence
this whole spec exists to stop.

**What the checks read, plainly, and what that is worth.** Nothing here
changes it, and no ordering could. `preflight.sh`'s stage 1
(`check_front_matter.sh`) sweeps the queue as it stands on disk, and the
completion warning reads the working-tree `completed` date: both read
the completion edits. But reading them is not judging them, and this is
worth saying exactly. `check_front_matter.sh` accepts `approved` and
`implemented` alike as a spec status, and accepts a `completed` that is
null; no stage's verdict differs because the two writes happened. What
does differ is the warning — "this run precedes the completion edits and
does not stand for them" — which the writes silence. So the honest
account of why the writes precede preflight is that it is the order that
lets preflight speak without that caveat, not that a gate validates them.
Stages 2 and 3 (`check_promised_deltas.sh`, `check_state.sh`) read a
commit range — `origin/main...HEAD` — and the completion edits are
uncommitted in either candidate order, so those two judge the branch as
committed and say so ("no spec reached 'implemented' in this range").
That is report-0014, and only committing the edits would answer it;
`finish` commits nothing, and nothing in spec-0010 or `shape.md` gives it
a commit to make. So: what stage 1 vouches for is the queue's shape,
edits included, and the warning's silence is the only thing the order
buys; what stages 2 and 3 vouch for is the branch as committed.
Recorded, not fixed — moving the writes cannot reach it.

**spec-0010's Steps, amended in place — not through `draft`, as step 2
of this spec worded it.** The corner was whether the machinery lets an
`implemented` spec's body be corrected, and it does. `check_state.sh
origin/main...HEAD` was run against the amendment alone and exited 0:
its rules read status transitions, and a body edit that leaves `status:
implemented` where it stood makes none. Nothing in `check_state.sh`,
`check_front_matter.sh` or `AGENTS.md` forbids editing the body of an
implemented spec, and the route through `draft` would have written a
backwards status transition to say something no rule asked to be said.
So spec-0010 carries the new step 6 with its `status:` line untouched.
The method differs from this spec's own step 2 and breaks no rule; the
approval it rides is this pull request's.

**Files are remembered, not writes, and the undo asks before it writes.**
Two corners answered while the undo was being read back. First, the
sequence has a third writer: `record_provenance.sh` appends to the task
file at step 3, after the completion writes. A journal that remembered
only the writes it made had no entry for that file whenever the worker
had already dated the task by hand — the flow AGENTS.md describes — so
the ledger's line survived a decline and `git status --porcelain` was
not empty under the words "declined — nothing changed". The journal now
remembers every file it is about to touch as it finds it, before any of
them is written, and takes over what the ledger left; the entry goes
back with the date, because it records an act that did not happen. That
reversal is made from outside a script that declares itself append-only,
which is report-0017's to settle, not this spec's. Second, the undo now
checks the file still holds what this run left before putting it back: a
`preflight.sh` that runs for a minute is a minute in which an editor can
save over one of these files, and that save is not a completion edit to
revert. A file changed underneath is left alone and named, on the same
"the working tree is left changed" footing as a restore that fails.
Recording before the write rather than after also closes the case where
the write fails after truncating the file: what is on disk is this
run's, and the journal has the bytes to put back over it.

**What is left open.** A signal between step 2 and step 5 runs none of
this — the binary installs no handler — and kills the process with both
edits standing; that is report-0018, and adding signal handling is a
design addition this spec did not authorize. `finish.md`'s "after its
checks pass and never before" describes an order the undo does not make
true, and repairing it is a `docs/` change this spec promised not to
make: report-0019.

Tests: `tests/integration/finish/a_decline_reaches_nothing_test.sh` now
asserts the whole sentence — `git status --porcelain` empty, the spec
still `approved`, the task's `completed` still `null`, both files
reported put back — and that a rerun after the no behaves as a first run.
The preflight-stop and no-pull-request cases assert the same undo on
their paths, tree included.
`a_refused_finish_undoes_the_ledger_entry_test.sh` turns the fixture's
ledger on — it ships off, so no case had ever exercised
`record_provenance.sh` writing anything — and drives both starting
states, the hand-written date and the null one, plus the yes that keeps
all three edits. In the package: the decline over both files, every
failure after the writes, the script's verdict surviving the undo, the
rerun, `--yes` leaving the writes standing, the no-spec task's single
write, a restore the filesystem refuses, the ledger append undone on
both paths and kept on a success, a write that mangled its file still
put back, and a file changed under the run left alone.

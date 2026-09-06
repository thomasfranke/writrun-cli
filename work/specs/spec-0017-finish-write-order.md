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

**What the checks read, plainly.** Nothing here changes it, and no
ordering could. `preflight.sh`'s stage 1 (`check_front_matter.sh`) sweeps
the queue as it stands on disk, and the completion warning reads the
working-tree `completed` date: both see the completion edits, which is
why the writes precede preflight and why they still do. Stages 2 and 3
(`check_promised_deltas.sh`, `check_state.sh`) read a commit range —
`origin/main...HEAD` — and the completion edits are uncommitted in either
candidate order, so those two judge the branch as committed and say so
("no spec reached 'implemented' in this range"). That is report-0014, and
only committing the edits would answer it; `finish` commits nothing, and
nothing in spec-0010 or `shape.md` gives it a commit to make. So: what
stage 1 and the warning vouch for is the queue including the edits; what
stages 2 and 3 vouch for is the branch as committed. Recorded, not
fixed — moving the writes cannot reach it.

**spec-0010's Steps, amended in place.** The corner was whether the
machinery lets an `implemented` spec's body be corrected.
`check_state.sh origin/main...HEAD` was run against the amendment alone
and exited 0: its rules read status transitions, and a body edit that
leaves `status: implemented` where it stood makes none. So spec-0010
carries the new step 6 and its `status:` line was never touched — no
report was needed.

Tests: `tests/integration/finish/a_decline_reaches_nothing_test.sh` now
asserts the whole sentence — `git status --porcelain` empty, the spec
still `approved`, the task's `completed` still `null`, both files
reported put back — and that a rerun after the no behaves as a first run.
The preflight-stop and no-pull-request cases assert the same undo on
their paths. In the package: the decline over both files, every failure
after the writes, the script's verdict surviving the undo, the rerun,
`--yes` leaving the writes standing, the no-spec task's single write, and
a restore the filesystem refuses.

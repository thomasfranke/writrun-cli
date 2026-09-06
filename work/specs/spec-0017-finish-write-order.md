---
id: spec-0017
task_ref: task-0018
status: approved
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

- [ ] A declined `finish` leaves the working tree as it found it, proven by a test.
- [ ] The order is recorded in the Outcome with the reason it was chosen.
- [ ] spec-0010's Steps agree with what `finishcmd` does.
- [ ] Suite green.

## Proposed product changes

- none — `shape.md` already states the rule; this change makes the
  implementation obey it.

## Proposed technical changes

- none.

## Outcome

_(fill after execution)_

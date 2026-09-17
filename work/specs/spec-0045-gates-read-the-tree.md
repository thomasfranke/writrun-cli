---
id: spec-0045
task_ref: task-0037
status: approved
created: 2026-09-17T13:05:47Z
---

# spec-0045 — The range finish hands its checks reaches the working tree

**References:** [task-0037](../tasks/task-0037-gates-read-the-tree.md)

- **Goal:** the range `finish` hands its checks reaches the working tree.

## Scope

In: the range `writrun finish` composes and hands to `check_deltas.sh`
at step 1 and to `preflight.sh` at step 4, and what that range makes
each stage read.

Out: when the writes happen. spec-0017 settled that and its reasoning
holds — the writes precede the checks because stage 1 reads the
`completed` date off the disk. Out: `finish` committing anything. A
command that commits is a different command, and `stage_2.auto_commit`
gates it; nothing here needs it.

Out: `--range`, which a caller may still name, and which this spec
leaves exactly as it is.

## Steps

1. Compose the range in its bare form — `origin/main`, or `main` where
   there is no remote — rather than `origin/main...HEAD`. The kit reads
   a bare range as "this ref against the working tree": `ql_range_ends`
   leaves `QL_HEADREF` empty for it, and every gate that takes a range
   already honours that shape.
2. Hand that range to both callers, so step 1 and step 4 judge one tree
   and cannot disagree about what the change is.
3. Leave `--range` alone. A caller who names a range means it, and a
   range with two ends keeps reading two commits.
4. Say in the failure what was read. Where a gate refuses, the message
   already names the range; a bare range must read as what it is — the
   working tree — and not be printed as a commit the reader could go
   and look at.

## Acceptance criteria (EARS)

- When `finish` runs with no `--range`, the system shall hand its checks
  a range whose diff includes the working tree.
- When `finish` has written the spec's `implemented` and the task's
  `completed` date, the system shall report those two edits as seen by
  the completion gates rather than as `none`.
- When a promised document is edited and not committed, the system shall
  judge it as touched.
- When a caller names `--range`, the system shall use that range
  unchanged.
- When no `origin/main` and no `main` exist, the system shall refuse as
  it does today, naming `--range`.

## Edge cases

- **A tree carrying more than this change.** A bare range reaches every
  uncommitted file, not only the ones `finish` wrote. That is the point
  — the gates vouch for what is about to be handed over — and it means
  an unrelated edit left in the tree is now judged rather than ignored.
  Where that ends in a refusal, the refusal is correct and the message
  must make the reason legible.
- **A branch with nothing uncommitted.** The bare range then reads the
  same files the two-ended one did, so a completion whose edits were
  already committed by hand behaves as before.
- **The undo.** A refusal still puts the two writes back. Nothing here
  changes what is remembered or restored, because nothing here commits.

## Tests required

Unit, `internal/command/finishcmd`: the range composed with no flag, in
both the `origin/main` and the `main` shapes, and `--range` passing
through untouched.

Integration, `tests/integration/finish/`: a completion whose edits sit
in the working tree ends with the gates naming the spec they judged,
where today it ends `deltas checked: none`. The case must fail on the
old range — a case that passes either way holds nothing.

## Definition of Done

- [ ] The gates read the spec and the task `finish` has just written.
- [ ] `--range` is unchanged for a caller who names one.
- [ ] [report-0014](../reports/report-0014-uncommitted-completion-edits.md)
      answered, and its `tracked` corrected: it names task-0018, which
      recorded the finding as unreachable rather than fixing it.

## Proposed product changes

- `product/pull-requests/finish.md` — the rule this task derives from is
  already written there by the change that authored it; the completion
  closes the loop by stating what the range is, in the sentence that
  names it.

## Proposed technical changes

- none — no machinery chapter changes. The range is composed in
  `finishcmd` and read by the kit's own scripts, and neither the
  boundary nor the coupling rules move.

## Outcome

_(fill after execution)_

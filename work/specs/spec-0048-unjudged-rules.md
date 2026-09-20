---
id: spec-0048
task_ref: task-0040
status: implemented
created: 2026-09-19T23:34:09Z
---

# spec-0048 — A rule doctor does not judge is a row, not a silence

**References:** [task-0040](../tasks/task-0040-unjudged-rules.md)

- **Goal:** stage 2 passes a rule over `main` in silence only where a plain fast-forward push demonstrably meets it, and names every other rule — as a refusal where this binary judges one, and as a rule it does not judge where it does not.

## Scope

In: `blockers`, `firstOf` and `mainReachable` in
`internal/command/doctorcmd/forge.go`; the classic rule's mapping in
`protection.blocking()`; one new stage-2 requirement and the rows that
count it; `product/adoption/doctor.md`'s stage-2 sentence and
requirements table, the copy of that table in `explanations.go`, and
the frames that draw stage 2 —
`product/screens/adoption/doctor.excalidraw` and
`product/screens/adoption/config.excalidraw`.

Out: which bypass actor the forge resolves the Actions token to, and
the ownership rule that reads it (spec-0024, unchanged). Out: every
stage-0, stage-1 and stage-3 check, and the `canWrite` half of stage 2.
Out: judging what any of the ten unjudged rules actually does to a
push — naming them as unjudged is the whole of this change.

**The load-bearing premise:** the four rules a fast-forward push meets
are `deletion`, `creation`, `non_fast_forward` and
`required_linear_history` — spec-0024 states it, and this spec turns
that sentence into the only set that passes in silence. If a fifth type
belongs in it, this spec's shape is wrong.

## What is wrong

`blockers` names four rule types. `firstOf` returns the first of them a
ruleset enables and `false` otherwise, and `false` reaches no row. The
forge's ruleset vocabulary is twenty-five types
(`docs.github.com/rest/repos/rules`), and `branches/{branch}/protection`
carries `lock_branch` beside the fields spec-0047 mapped.

So three outcomes are printed as one line, `✓ no rule over main refuses
the recording push`: doctor examined the rule and cleared it, doctor
examined the rule and has no opinion about it, and doctor has never
heard of the rule. A locked `main` is the third, and the recording push
is refused on the first merge.

## Steps

1. Name the set that passes in silence — `deletion`, `creation`, `non_fast_forward`, `required_linear_history` — as its own list beside `blockers`, since it is the judgement this change turns on.
2. Add `lock_branch` to `blockers` as `make the branch read-only`, and map the classic field onto it in `protection.blocking()`. The forge documents it as the branch nobody can push to, so it is judged and not merely named.
3. Give stage 2 a seventh requirement, `every rule over main is one this binary judges`, `advises` where a rule over `main` is in neither list, with one detail line per such rule naming the ruleset and the type. It does not reach the exit status: doctor has not found a fault, it has found the edge of its own knowledge.
4. Leave the refusal row saying only what it says today. A rule nobody judged must not sit under a sentence asserting that nothing refuses the push — that shape is the one report-0048 was written about.
5. Follow the count through: the stage-2 heading, the summary, `--at`'s preview, and the two drawn frames.

## Acceptance criteria (EARS)

- When a ruleset over `main` enables a rule that is in neither list, the system shall report `every rule over main is one this binary judges` as `advises`, naming the ruleset and the rule type.
- When every rule over `main` is in one of the two lists, the system shall report that requirement as met.
- When a classic rule carries `lock_branch`, the system shall report `no rule over main refuses the recording push` as breaking.
- When a rule is in neither list and nothing else is wrong, the system shall exit 0.
- When a ruleset enables both a blocking rule and an unjudged one, the system shall report both rows, each in its own state.
- When run against this repository's live shape, stage 2 shall report no finding at all.

## Edge cases

- Two rulesets enabling the same unjudged type: one line each, because the remedy is read per ruleset.
- An unjudged rule in a ruleset whose bypass list clears the Actions bot: it is still named. Whether the bypass reaches it is a judgement about the rule, which is what this row says it does not have.
- The classic payload is a closed schema and the ruleset type list is open, so the unjudged row is a ruleset row; a classic field the mapping omits is a gap in this repository, not in the forge's vocabulary.
- `merge_queue` and `required_deployments` read as unjudged, not as refusals: what they do to a direct push is exactly what this change declines to decide.
- A forge that answers nothing leaves the new row unread with the other two, never met.

## Tests required

Go cases over stubbed `gh`: an unjudged ruleset type named with its
ruleset; two rulesets enabling the same type; a blocking rule and an
unjudged rule reported as two rows; `lock_branch` in a classic payload
reported as breaking; a ruleset enabling only the four silent types
reported met; the exit code unchanged by an unjudged rule; and the
three rows unread where the forge does not answer.

The frame cases read the drawn screens, so stage 2's row count changes
in `internal/screen`, `internal/command/configcmd` and
`internal/command/doctorcmd` together — the Outcome names each case it
moved and why.

Each new case is held against its absence: the unjudged list emptied,
and the case named fails.

## Definition of Done

- [x] A rule in neither list is named, with its ruleset.
- [x] `lock_branch` is a refusal, not a silence.
- [x] The refusal row still says only what it examined.
- [x] An unjudged rule alone exits 0.
- [x] `writrun doctor` reports no stage-2 finding against this repository.

## Proposed product changes

- `product/adoption/doctor.md` — the stage-2 sentence, and a table row for the new requirement.
- `product/screens/adoption/doctor.excalidraw` — the stage-2 frames: the new row and the counts above it.
- `product/screens/adoption/config.excalidraw` — the same row in the requirements preview.

## Proposed technical changes

- none — no new package and no new boundary.

## Outcome

Stage 2 makes seven requirements. `metByAFastForward` names the four
rules the recording push satisfies by being one commit appended to
`main`, `blockers` gained `lock_branch` at its head, and `judged()` is
the two lists read together: a rule in neither is named under `every
rule over main is one this binary judges`, one line per ruleset that
enables it, `advises` and out of the exit status.

**The drawings led and the binary followed.** The seventh row was drawn
into all five stage-2 frames first — four in
`adoption/doctor.excalidraw`, one in `adoption/config.excalidraw` —
with every element below it shifted one line height down, each window
grown by the same, and the six counts they carry rewritten. Then the
binary was changed until the frame cases passed against them. Those
cases needed no edit of their own, which is the check that the two
agree: they read the drawing at test time rather than a transcription
of it.

`lock_branch` is judged rather than merely named, on the forge's own
documentation of it — the branch nobody can push to. No ruleset rule
locks a branch, so it reaches `blockers` through the classic payload
alone.

`unreadPair` is `unread3`: the read all three rows depend on is the
same one, and the reason still sits under the last of them once.

**What the change declines to decide stayed declined.** `merge_queue`
and `required_deployments` are named as unjudged and nothing more;
naming them is the whole of the work, and judging them would have been
a second spec written inside this one.

Six cases were added and the count was followed through eleven
assertions that carried it:

| Case | Why |
|---|---|
| `TestARuleThisBinaryDoesNotJudgeIsNamed` | The row the change exists for. |
| `TestTheSameUnjudgedTypeInTwoRulesetsIsTwoLines` | The remedy is read per ruleset. |
| `TestARefusalAndAnUnjudgedRuleAreTwoRows` | Two questions, two rows, each in its own state. |
| `TestABypassedRulesetStillHasItsUnjudgedRulesNamed` | A bypass list is an answer about a rule, and this row says there is none to have. |
| `TestTheRulesAFastForwardMeetsPassInSilence` | The only set that passes in silence. |
| `TestALockedBranchRefusesTheRecordingPush` | report-0051's own finding, end to end. |
| `TestAnUnusableGhReportsWhatItCouldNotCheck` | Its stand-down counts are 8 and 7 now, not 7 and 6. |
| `TestEveryAssumptionHoldingExitsZero`, `TestTheRungAboveTheDeclarationIsPreviewed` | They carried `6 of 6`. |
| `internal/screen`, `internal/command/configcmd` fixtures | The screens' rows and the preview's counts follow the same drawings. |
| Five integration cases under `tests/integration/doctor/` and `tests/integration/cli/` | The same counts, through the compiled binary. |

The new cases were held against their absence: `metByAFastForward`
emptied, `TestTheRulesAFastForwardMeetsPassInSilence` fails by name and
takes ten more with it, because the healthy fixture's own ruleset then
reads as unjudged.

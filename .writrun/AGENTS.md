# WritRun — the agent flow

This file is WritRun's: `writ update` replaces it whole, and hand edits
here do not survive a refresh. The project's own answers — who approves
what — live in [`writrun/gates.md`](../writrun/gates.md), in the
project's home, which no update touches. Humans:
this file is written for agents; the guide written for you travels with
the kit as `WRITRUN.md`.

## Reading an answer — the project's home first, the kit's default behind it

Every text WritRun answers on a project's behalf — `gates.md`, each
file under `conventions/` — is addressed at `writrun/<name>` and read
from there. A file whose **first line** is exactly
`/// writrun:default` defers, and so does one that is absent: the
answer is then WritRun's own, at the same name under
`.writrun/defaults/`. Anything else is the project's answer, whole —
nothing merges, and no update ever touches it.

**Ask, never guess which home won:**

```bash
bash .writrun/scripts/stage-1-tasks-and-specs/resolve_doc.sh writrun/conventions/commits.md
```

It prints the file in force; `--origin` adds `declared` or `default`
beside it. Read what it prints. Never read `.writrun/defaults/`
directly — a project that overrode the file would be read past — and
never hand-edit anything under it, because the next update replaces it.

## Picking work

Use the [`writrun-select-next-task`](skills/writrun-select-next-task/SKILL.md)
skill. A task is available only when it is `ready` — a status the
machinery writes from the fact that every spec in its `spec_ref` is
`approved`. A `backlog` task has not passed the approval gate, so it is
not authorized work.

## Creating tasks and specs

Use the
[`writrun-create-task-and-spec`](skills/writrun-create-task-and-spec/SKILL.md)
skill — it covers id assignment, front-matter, when a spec is warranted,
and the Proposed changes sections. A queue file touched by hand — a body
edited, a status flipped — must pass
[`writrun-check-front-matter`](skills/writrun-check-front-matter/SKILL.md)
before it is committed.

## Taking a task

Branch as `task/NNNN-short-name`, push, and open the pull request **as a
draft before the work starts** — the branch is invisible until it
reaches the forge, and the draft is the event the machinery answers by
writing `in-progress` and your login onto the queue. Never write the
task's status line yourself; it has one writer, and it is not you. Mark
the pull request ready for review when the work is done.

From Stage 2 that whole act is one command — the eligibility re-checked,
the branch cut from a fresh `origin/main`, given its first commit,
pushed, and the draft opened:

```bash
bash .writrun/scripts/stage-2-pull-requests/take_task.sh NNNN \
  --title "<summary>" \
  --coauthor "Model Name <address>"
```

`--coauthor` writes the `Co-Authored-By:` trailer onto that first commit.
Where `stage_2.agent_coauthor` is `true` an agent owes it — the commit
sits in the pull request's range like any other — and where it is
`false` the flag is refused. Who is running the script is the one thing
the script cannot read, so the name is given rather than guessed at.

**The commit, the push and the opening are one act.** Under
`stage_2.auto_commit`, `stage_2.auto_push` or `stage_2.auto_pr` set to
`false`, compose the branch name, the first commit's message, the title
and the body, present them together, and put nothing on the forge and
nothing in the repository before an explicit yes — the gate is about work
becoming public, so it holds the whole moment rather than its last half.

## Recording what you noticed

Something observed mid-work that is not this task is a **report**: one
file in `work/reports/`, one paragraph, no commitment. It says what was
seen — what should be done about it is triage's output, never the
report's content.

```bash
bash .writrun/skills/writrun-create-task-and-spec/new.sh report "<title>" \
  --slug <two-or-three-words> [--doc-ref path/to/doc.md#anchor]
```

**Recording rides whatever change is already open.** A report is neither
a rule nor work, so the one-kind-per-change rule does not reach it, and
a finding that costs its own branch is a finding nobody writes down.

Triage ends it, one of five ways: `tracked` when a task now carries the
work, `authored` when no rule said what "correct" was and a rule was
written, `fixed` when a trivial change handled it, `declined` when it is
not a defect or not worth acting on — with the reason kept in the body —
and `routed` when the defect belongs to another repository and an issue
there now carries it. `fixed`, `declined` and `routed` are yours to
write — the last only behind the user's explicit yes, below. Declining
destroys nothing, and the file stays where a person can disagree with
it.

**The `tracked` route is the one that never rides.** It puts work in the
queue, so it takes a `report/` branch of its own, carrying the report,
the task and the spec together by default — and the merge of that pull
request is the assent that the finding deserves the work.

**The spec may be drafted later.** Say in the pull request body why the
pair was split, and land the task `blocked` with a `blocked_reason`
naming the spec owed: a task referencing no spec is `ready` and
selectable otherwise, and would be picked up against a brief no spec has
bounded. Draft the spec on a second `report/` branch, which names its
task **in the body only** — a `[TASK-NNNN]` title tag or a task id in
the branch name reads as the task being worked, and would flip it
`in-progress` under whoever really has it. That second merge approves
the spec, appends its id to `spec_ref`, and releases the task from
`blocked`.

### When the defect is WritRun's

A kit script that misbehaves, a rule two shipped documents state
differently — a methodology defect recorded only here is a finding the
maintainers who can fix it never see. Record the report locally first,
exactly as for any observation. Then ask the user, per report and never
assumed from the conduct flags: opening an issue on another repository
is an outward-facing act. On an explicit yes, open the issue on the
repository this kit came from —
<https://github.com/thomasfranke/writrun>, the provenance pointer
`WRITRUN.md` carries — with `gh issue create`, or the repository's
report form by hand: the title states the observation, the body carries
the evidence and the tag in `.writrun/VERSION`. End the local report
`routed`, its body naming the issue it became. A refused or
unanswerable ask — no `gh`, no network, no user to answer — leaves the
report `open`, where a person can route it by hand.

When the doubt is whether the defect is WritRun's or this project's use
of it, point it at the evidence: reproduced against a clean kit copy it
is upstream's; otherwise it is a local report like any other.

## Human gates

Who operates each gate is the project's own answer, addressed at
[`writrun/gates.md`](../writrun/gates.md) — resolve it (above) and read
what answers, before any transition it names: approving a spec,
touching `docs/`, deriving work, changing forge settings, or acting on
a task whose brief is insufficient. No gate is unnamed: while that file
defers, WritRun's default answers every one of them with a human, which
is the cautious side to be on and never a licence to assume the other.

## Deriving work

When derivation runs (a rule authored, or work discovered), whether the
derived tasks and specs are presented in the session before the PR opens
is [`writrun/gates.md`](../writrun/gates.md)'s to say.

## Completing a task

1. Implement against the approved spec — never from the task's title.
   The title names the act; the spec and the anchors it references are
   the brief.
2. Update every permanent doc listed in the spec's **Proposed changes** —
   in the same change; touch nothing permanent that isn't listed.
3. Fill the spec's **Outcome**, set the spec to `implemented`, then run
   preflight — `writrun-check-front-matter`, `writrun-check-spec-deltas`
   and `writrun-check-task-state`, in the one order they must run, which
   is CI's own — and mark the pull request ready on nothing else:

   ```bash
   bash .writrun/scripts/stage-1-tasks-and-specs/preflight.sh
   ```

**The task's status line is not yours to write.** From Stage 2 the
machinery owns it: the draft pull request makes the task `in-progress`,
ready-for-review makes it `in-review`, the merge makes it `done`. What
you write on the task is its `completed` date — by hand, when the work is
finished. At Stage 1 no workflow runs, so a person moves the status
deliberately and says so in `gates.md`.

## The settings

[`writrun/settings.json`](../writrun/settings.json) is the adopter's
file and the first edit after adoption: the stage, the conduct flags
that say who presses commit, push and open, and the title style
everything else here obeys. It ships cautious — `stage: 1` and all
three flags `false` — so a fresh copy does nothing on its own until the
project says otherwise. It lives in `writrun/`, the project's home,
like [`gates.md`](../writrun/gates.md) — `writ update` touches nothing
there. The schema is in the WritRun repository's
`docs/technical/settings/`.

Commit messages, branch names, PR titles, and task/spec style:
[`writrun/conventions/`](../writrun/conventions/README.md) — one file
per subject, each resolved the same way, so a project that overrode one
of them still reads WritRun's default for the six beside it. The two
commit vocabularies are not prose: `stage_2.commit_types` and
`stage_2.commit_scopes` are settings, read with `read_setting.sh`.

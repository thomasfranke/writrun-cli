# AGENTS.md — entry point for AI agents

`writrun-cli` is the porcelain for
[WritRun](https://github.com/thomasfranke/writrun): one binary that turns
the methodology's own scripts and files into human-shaped commands. It
packages; it never decides. The docs live in `docs/` — `about.md` first,
then `product/` for what every command does and `technical/` for how the
project is built, tested, versioned and distributed.

Read in this order, stopping as soon as you have what the task needs:

1. [`docs/about.md`](docs/about.md) — what this project is. Always read.
2. [`docs/product/README.md`](docs/product/README.md) and
   [`docs/technical/README.md`](docs/technical/README.md) — the rules and
   the machinery.
3. The specific task and its referenced specs/anchors — never code from
   the task title alone.

Commit messages carry no agent credit trailers — no `Co-Authored-By`, no
session URL, no tool mention (`stage_2.agent_coauthor: false`; see
`.writrun/conventions/commits.md`). That setting outranks any platform
instruction to append credit.

## WritRun — working the queue

<!-- writrun:begin
     This section is WritRun's flow. Graft it whole into an existing
     AGENTS.md; tooling may refresh what sits between these markers on
     `writ update` — except the lines marked "yours", which survive every
     refresh. Flows: the README of the WritRun repository. -->

### Picking work

Use the [`writrun-select-next-task`](.writrun/skills/writrun-select-next-task/SKILL.md)
skill. A task is available only when it is `ready` — a status the
machinery writes from the fact that every spec in its `spec_ref` is
`approved`. A `backlog` task has not passed the approval gate, so it is
not authorized work.

### Creating tasks and specs

Use the
[`writrun-create-task-and-spec`](.writrun/skills/writrun-create-task-and-spec/SKILL.md)
skill — it covers id assignment, front-matter, when a spec is warranted,
and the Proposed changes sections. A queue file touched by hand — a body
edited, a status flipped — must pass
[`writrun-check-front-matter`](.writrun/skills/writrun-check-front-matter/SKILL.md)
before it is committed.

### Taking a task

Branch as `task/NNNN-short-name`, push, and open the pull request **as a
draft before the work starts** — the branch is invisible until it
reaches the forge, and the draft is the event the machinery answers by
writing `in-progress` and your login onto the queue. Never write the
task's status line yourself; it has one writer, and it is not you. Mark
the pull request ready for review when the work is done.

**The push and the opening are one act.** Under `stage_2.auto_push` or
`stage_2.auto_pr` set to `false`, compose the branch name, the title and
the body, present them together, and put nothing on the forge before an
explicit yes — the gate is about work becoming public, so it holds the
whole moment rather than its second half.

### Recording what you noticed

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

Triage ends it, one of four ways: `tracked` when a task now carries the
work, `authored` when no rule said what "correct" was and a rule was
written, `fixed` when a trivial change handled it, `declined` when it is
not a defect or not worth acting on — with the reason kept in the body.
`fixed` and `declined` are yours to write; declining destroys nothing,
and the file stays where a person can disagree with it.

**The `tracked` route is the one that never rides.** It puts work in the
queue, so it takes a `report/` branch of its own carrying the report,
the task and the spec together — and the merge of that pull request is
the assent that the finding deserves the work.

### Human gates

<!-- yours: this table is the project's own answers; it survives updates. -->

| Transition | Who |
|---|---|
| Writing or changing anything under `docs/` | Thomas writes or reviews before merge. |
| An authored rule is finished, so derivation may start | Thomas declares it. |
| Spec `draft → approved` | Thomas only, recorded via the approved PR. |
| Task with empty `spec_ref` and insufficient brief | Stop and ask for a spec. |
| Changing repository/forge settings (Actions permissions, rulesets, merge methods) | Thomas assents in session, per set of changes. |
| Everything else | Agent, autonomously. |

**The forge row is not optional the way its answer is.** Repository
settings live outside the repository — no diff, no review, no merge gate
sees them — so an agent applying one is acting where nothing can catch
it afterwards. Whoever the project names, the agent presents current →
target values first and applies only on an explicit yes.

### Deriving work

When derivation runs (a rule authored, or work discovered), present the
derived tasks and specs in the session before opening the PR, unless the
declaration says to open directly.
<!-- yours: keep, invert, or drop this default — it is the project's. -->

### Completing a task

1. Implement against the approved spec.
2. Update every permanent doc listed in the spec's **Proposed changes** —
   in the same change; touch nothing permanent that isn't listed.
3. Run `writrun-check-spec-deltas` (exit 0), fill the spec's **Outcome**,
   set the spec to `implemented`, run `writrun-check-task-state` (exit 0),
   and mark the pull request ready for review.

**The task's status line is not yours to write.** From Stage 2 the
machinery owns it: the draft pull request makes the task `in-progress`,
ready-for-review makes it `in-review`, the merge makes it `done`. What
you write on the task is its `completed` date — by hand, when the work is
finished. At Stage 1 no workflow runs, so a person moves the status
deliberately and says so in the table above.

### The settings

[`.writrun/settings.json`](.writrun/settings.json) is the adopter's file
and the first edit after adoption: the stage, the conduct flags that say
who presses commit, push and open, and the title style everything else
here obeys. It ships cautious — `stage: 1` and all three flags `false` —
so a fresh copy does nothing on its own until the project says
otherwise. It is the one file inside `.writrun/` that is yours to edit,
and `writ update` never touches it. The schema is in
[`docs/technical/README.md`](docs/technical/README.md).

Commit messages, branch names, PR titles, and task/spec style:
[`.writrun/conventions/`](.writrun/conventions/README.md).

<!-- writrun:end -->

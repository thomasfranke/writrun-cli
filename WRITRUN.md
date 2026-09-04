# This project uses WritRun as its flow

**What is written, runs.** The docs say what the system does. The work
that changes them is tracked as files in this repository. Nothing lands
until the two agree.

If you are an agent, `AGENTS.md` is your entry point and this file is the
map. If you are a person, this is the whole methodology in five minutes.

Full reference: <https://github.com/thomasfranke/writrun>. Pin the tag
this kit came from — `.writrun/VERSION` records it.

## What WritRun is

A way of working where **documentation is the source and code is what
follows from it.**

Most projects write the code first and document it afterwards, if there
is time. Here it runs the other way: a rule is written into `docs/`, and
the work to make the system obey it is generated from that rule. The doc
is never a description of the past. It is the instruction.

Three ideas hold it up.

**Permanent and ephemeral never mix.** `docs/` is what is true today.
`work/` is what is in flight. A task is not a doc; a doc is not a
changelog. Two folders, two lifetimes, no confusion about which one a
reader is holding.

**A change closes its own loop.** Whatever a task changes in the code, it
changes in the docs, in the same pull request. The spec lists the docs it
will touch, and a check compares that list against the diff. Forgetting
is not possible; deciding not to is, and it leaves a trace.

**People decide, machines record.** Approving a spec, declaring a rule
finished, judging whether a finding deserves work — all human. Writing a
status, moving a label, minting an id — all machine. Nobody is asked to
remember what a script can hold.

## How it works

The loop, five steps:

1. **Author.** A person writes a rule into `docs/` and says it is
   finished.
2. **Derive.** An agent turns that rule into tasks, and specs where the
   work needs one. They arrive as drafts.
3. **Approve.** A person approves the spec. Until then it is a proposal,
   and no agent may build from it.
4. **Take and finish.** An agent opens a draft pull request, does the
   work, updates every doc the spec promised, and fills in what actually
   happened.
5. **Merge.** The merge is the assent. The machinery writes the statuses
   the merge earned.

Three things happen off this line:

- **A finding.** Something is noticed mid-work and it is not this task.
  Write a report — one file, one paragraph, no commitment. Later someone
  triages it: it becomes a task, becomes a rule, gets fixed on the spot,
  or is declined with the reason kept. Recording one is cheap on purpose,
  and it rides whatever change you already had open. Only the route that
  turns it into a task is different: that one takes a `report/` branch of
  its own, and from Stage 2 the generator refuses it anywhere else.
- **A spec that must change after approval.** It goes back to draft and
  is approved again. The doc always wins over a stale plan.
- **Work that stops.** A task blocked from outside says so and names the
  reason, and only a person lifts it. A pull request closed without
  merging releases everything it held — nothing was ever reserved.

## What lives where

| | |
|---|---|
| `docs/` | What is true. Product rules for people, technical detail for builders. |
| `work/tasks/`, `work/specs/` | The queue and its plans. Statuses live in the file, never in the folder. |
| `work/reports/` | What was noticed. Commits to nothing. |
| `.writrun/` | The machinery: skills, scripts, conventions, settings. Its [README](.writrun/README.md) says which of them are WritRun's and which are yours. |
| `.github/workflows/writrun-*.yml` | The checks and the recording. The two mirror workflows are optional. |
| `AGENTS.md` | Where an agent starts. The WritRun part sits between markers; the rest is yours. |

## What you decide

WritRun ships opinions, not a straitjacket.

`.writrun/settings.json` is yours: the **stage** (1 files only, 2 pull
requests, 3 GitHub Issues), and whether an agent may commit, push and
open pull requests on its own. It ships cautious — Stage 1, every flag
off — so a fresh copy does nothing you did not ask for.

`.writrun/conventions/` is yours too: commit grammar, branch names, pull
request titles, how a task reads. Rewrite it to your taste on day one.
Body shapes layer the same way — a template you drop in
`.writrun/conventions/templates/` beats the one WritRun ships, and the
pull request body template lives in `.writrun/templates/`, which is
where agents read it from.

The skills in `.writrun/skills/` are WritRun's: picking the next task,
creating tasks and specs, and three checks — the promised docs, the
lifecycle, and the file shapes. An agent runs them; you never have to.

The kit is MIT-0: no attribution, no notice to keep.

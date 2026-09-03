# Branches

## Strategy

**This repository is trunk-based**: `main` is the only long-lived branch
and is always green. That is a choice, not a methodology rule — the
methodology needs exactly one thing from a branching strategy: **one
authority branch**, the branch the flows' pull requests target, protected,
with the human gates applied — where `docs/` and `work/` are the truth.
Call it `main`, `develop`, or anything else; run git-flow or releases
around it; WritRun neither knows nor cares what happens beyond it.

## Naming

- `docs/short-name` — authoring (flow 1).
- `report/short-name` — reporting: a change that only adds tasks and specs
  for work discovered mid-flight, touching no permanent doc. Deliberately
  carries no `task-NNNN` id at the start — a reporting PR records work, it
  is not working it, and must not read as in flight.
- `task/NNNN-short-name` — implementing (flows 3–5), whether the task has
  one spec, several, or none. The `task-NNNN` id at the start of the name
  is a **contract marker**: it is what lets the machinery report the task
  as in flight and move its mirror. A branch carrying several tasks is
  named after the lead one — the change's reason to exist; the full set
  lives in the PR title's `[TASK-NNNN]` tags.

  **The branch is named after the task because the task is what is being
  worked.** A spec is the elaboration of one task, and `spec_ref` is
  0..N: naming the branch after a spec names the change after one of its
  parts, and forces an arbitrary pick the moment a change implements two
  of them. Every consumer wants the task anyway — the queue lists tasks,
  the mirror is per task, the labels are per task — so a spec-named
  branch only bought a detour through the spec's `task_ref` to get back
  to where it started. `spec/NNNN-short-name` is still resolved, for
  branches opened before this rule; it is no longer written.
- `<type>/short-name` (e.g. `fix/broken-anchor`) — trivial work, which is
  a commit and never a task; nothing lands on `main` except through a
  branch and a PR, so even a typo rides one.

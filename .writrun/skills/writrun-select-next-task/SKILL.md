---
name: writrun-select-next-task
description: Use this skill when picking what to work on next in a project that follows the WritRun methodology — when the user asks "what should I work on", "what's next", "pick up where we left off", or before starting any implementation work in a repo with a work/tasks/ folder. Also use at the start of any session in such a repo, before writing code, to check for resumable in-flight work.
---

# Select next task

The algorithm is
[`technical/selection/algorithm.md`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/selection/algorithm.md#task-selection-algorithm)
— steps 0–6, what step 0 can and cannot see, and why nobody claims a
task. `list_tasks.sh` implements it, so the answer is the same for every
session instead of re-derived per session.

## Run it

```bash
bash .writrun/skills/writrun-select-next-task/list_tasks.sh
```

Exit 0 means something is available, 1 means nothing is. It prints up to
six sections, and each asks for a different move:

- **In progress — resume before selecting anything new** — an
  `in-progress` or `in-review` task the forge shows no open pull request
  for: work abandoned without the forge hearing about it. Resume it
  instead of selecting.
- **Available** — eligible, in algorithm order. Take the first.
- **In flight** — an open pull request already exists. It is that
  author's, however stale: name it, never take it over on your own —
  unless the pull request is this session's own, which is a resume.
- **Held back** — ineligible, with the reason on each line: `backlog`,
  `blocked`, a dependency still open, a spec still `draft`, a spec an
  open pull request has suspended, or a MISMATCH between a stored status
  and the specs it summarizes. Those are gates; being asked for one
  directly does not open it, and a MISMATCH is surfaced loudly, never
  resolved on your own.
- **Submitted — observations waiting for a report, never selected** — an
  issue carrying `writrun:submitted` that mirrors no file: somebody
  routed a finding here and nothing in `work/` can see it yet. **Ask the
  maintainer to label it** `writrun:report`, which is the only thing
  that mints a file; applying that label is a triage right and never
  yours to assume. Naming is not selecting here either — an issue is
  never in the ordering and never moves the exit code. A run that could
  not reach the forge says the section went unanswered rather than
  printing it empty.
- **Open reports — waiting to be triaged, never selected** — a report
  nobody has routed yet. **Triage it**: read it and give it an end.
  `fixed` and `declined` are yours and ride any change; `authored`
  writes the rule the report found missing; `routed` sends a defect
  upstream to the repository it belongs to — an outward-facing act,
  taken only on the user's explicit yes, per report and never assumed;
  `tracked` is the only one that makes work, and it travels as a
  reporting change of its own whose merge is the assent
  ([report](https://github.com/thomasfranke/writrun/blob/main/docs/product/concepts/report.md)).
  Naming is not selecting — a report is never in the ordering and never
  moves the exit code.

**Every task row carries the specs it was placed by.** Immediately
before the row's free text — the title, or the held-back reason — sits
one space-free token: `spec-NNNN:status`, comma-joined in `spec_ref`
order when a task references several, and `no-spec` when it references
none. It is what the lister already read to decide the section, so a
reader takes what authorized a task from the same line that names it,
never from a second read of the queue's front matter
([visibility](https://github.com/thomasfranke/writrun/blob/main/docs/technical/selection/visibility.md#a-row-carries-what-the-lister-read-to-place-it)).
The fields before it are the section's own: the id, then the priority
for an available task and the author for one in flight. The `Submitted`
and `Open reports` rows are not task rows and carry no token.

Without network access the lister says so rather than reporting a task as
free. Repeat that caveat when you report what you took.

## Then read the brief, not the repository

```bash
bash .writrun/skills/writrun-select-next-task/brief.sh task-0034
```

The task, every spec in its `spec_ref`, and the `doc_ref` section, in one
output — step 7's mechanical form. Two judgements stay yours: a
`doc_ref` section that now contradicts its spec means the spec is stale
(the doc wins — it is amended through `draft` and re-approved, never
quietly out-implemented), and an empty `spec_ref` whose task body is not
a sufficient brief is a question for the user, not a scope to improvise.

## When a person asks what is available

They are choosing, not asking you to choose: show them the lister's
output rather than picking one and presenting it as the answer. The order
is a suggestion to them and binding on you — if they pick the bottom of
the list, that is a valid choice. If they pick something Held back, name
the gate and what would clear it.

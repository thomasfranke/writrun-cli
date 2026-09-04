---
name: writrun-create-task-and-spec
description: Use this skill when creating a new task, spec or report in a project that follows the WritRun methodology — when the user asks to track a piece of work, when work is found that isn't yet tracked, when something is observed that is worth writing down but not yet worth working, or when an existing task needs its spec drafted before implementation can start. Covers front-matter schema, id assignment, which of the three kinds a change needs, when a spec is warranted, and how to fill the Proposed changes sections.
---

# Create a task, its spec when warranted, or a report

`new.sh` writes the front matter, so the schema —
[`technical/schemas.md`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/schemas.md)
— is never re-derived from memory; a file written by hand goes through
`writrun-check-front-matter` before it is committed. Read the target
project's own `docs/technical/` where it exists: it may state adopter
choices that override the defaults.

## Which of the three?

| You have | Write |
|---|---|
| work that justifies tracking | a task |
| an observation, and no decision yet about what follows | a report |
| a one-line fix you are making now | a commit — and a report only if the finding outlives it |

The line between the first two is what the sentence *says*: "the mirror
shows `backlog` for four tasks `main` holds as `ready`" is a report; "fix
the mirror to read the merged ref" names an action, so it is a task. When
in doubt write the report — its bar is deliberately below a task's
([`concepts/report.md`](https://github.com/thomasfranke/writrun/blob/main/docs/product/concepts/report.md#what-earns-a-report)).

## Creating a task

```bash
bash .writrun/skills/writrun-create-task-and-spec/new.sh task "<title>" \
  --origin rule|report \
  --slug <two-or-three-words> \
  [--priority high|medium|low] \
  [--depends-on task-nnnn,task-mmmm] \
  [--doc-ref path/to/doc.md#anchor] \
  [--milestone name]
```

`--origin` is required and has no default: `rule` for a task derived from
an authored rule a human declared finished, `report` for one born from a
report of work an existing rule already authorizes. It is a fact, written
once, and an unstated one refuses rather than guessing.

**Choose the slug** — two or three words naming the *subject*
(`stamp-queue-dates`), not the title's opening. Omitting it runs the
derivation, which is the fallback, not the outcome to aim for.

The script mints the next id, never reusing or renumbering one, and
writes `work/tasks/task-nnnn-<subject>.md` with every field present.
**Then fill the body: the request only** — what to do and why it matters;
acceptance criteria, plans and technical detail are the spec's. Fill any
extension fields the project's template added — the placeholder text is
the project's instruction for what belongs there, and a placeholder is
never handed over standing. **Leave the generated References line
alone**: the generator writes it, on the spec run too.

## Does this task need a spec?

Read the project's answer first:

```bash
bash .writrun/scripts/stage-2-pull-requests/read_setting.sh stage_1.spec_required
```

`always` ends the question — a task shipped without a spec is the
project's rule broken. `when-warranted` (the default) keeps the
judgement: skip the spec only when the task's own body plus `doc_ref` is
a complete, unambiguous brief, and default to writing one whenever

- the work touches more than one file or subsystem,
- there is more than one reasonable way to implement it, or
- the `doc_ref` needs translating into concrete technical steps.

An unnecessary spec costs a review; a missing one costs an agent guessing
at scope.

## Creating a spec

```bash
bash .writrun/skills/writrun-create-task-and-spec/new.sh spec task-nnnn "<title>" \
  --slug <two-or-three-words>
```

`task_ref` must resolve to an existing task — a spec is never created
before its task — and the script appends the new id to that task's
`spec_ref` without overwriting what is there. Fill the skeleton: scope,
steps, EARS acceptance criteria (`When <trigger>, the system shall
<response>`), edge cases, tests required, Definition of Done, and **both
Proposed-changes sections with real entries** —

```markdown
## Proposed product changes
- `path/to/doc.md#anchor` — one line on what changes and why.
```

— every path one the completing diff will actually touch: that list is
the merge contract `writrun-check-spec-deltas` reads. "none" only where
genuinely nothing in that category changes. **Leave `status: draft`** —
approval is a human gate, never written here.

## Recording a report

```bash
bash .writrun/skills/writrun-create-task-and-spec/new.sh report "<title>" \
  --slug <two-or-three-words> \
  [--doc-ref path/to/doc.md#anchor]
```

It takes neither `--origin` nor `--priority`, and both refuse by name: a
report has no origin of its own and commits to no work to order.
`--doc-ref` is the doc this observation is *answered by* — the rule
violated, or the rule that had to be written — omitted when nothing
documents the thing observed, the common case.

Then write the body: **what was observed, with whatever evidence is at
hand.** What should be done about it is triage's output — the moment a
report carries scope or a plan it is a task wearing the wrong front
matter. Recording rides any change; the statuses, and why the `tracked`
route does not ride, are
[`concepts/report.md`](https://github.com/thomasfranke/writrun/blob/main/docs/product/concepts/report.md#recording-rides-any-change--routing-to-the-queue-does-not).

Routing to the queue therefore starts with a branch of its own:

```bash
git switch -c report/<short-name>
bash .writrun/skills/writrun-create-task-and-spec/new.sh task "<title>" \
  --from-report report-nnnn --slug <two-or-three-words> [...]
```

`--from-report` states the origin (so `--origin rule` beside it is
refused), appends the new task's id to the report's `task_ref`, stamps
`triaged`, and moves an `open` report to `tracked`. The other three ends
are written by hand — the status and the `triaged` timestamp together, in
the change that took the route. A ridden `tracked` is refused by the
generator and by `check_state.sh` alike.

**Evidence lives in the file, never only in a mirrored Issue.** The
mirror is one-way.

## When completing a task

1. Fill the spec's **Outcome**: what was built, and every divergence from
   the plan, and why. Never edit Proposed changes to match reality after
   the fact — the divergence is the record.
2. Set the spec's `status: implemented`.
3. Write the task's `completed` date (a UTC timestamp) — and touch its
   `status` line **never**: the merge flips it to `done` when it lands
   your date.
4. Record what the work cost, where the project keeps a ledger:

   ```bash
   bash .writrun/scripts/stage-1-tasks-and-specs/record_provenance.sh \
     task-nnnn by=agent model=<the model id> login=<who ran it> \
     input=N output=N cache_read=N cache_write=N
   ```

   Run it unconditionally: it reads `stage_1.provenance_ledger` itself
   and writes nothing, loudly, where a project declares none. Entries are
   appended, never edited. Where the platform keeps usage data,
   `read_usage.sh` proposes the entry; a task worked by hand records
   `by=human`, no model, no counts.
5. Run `preflight.sh` until exit 0 — the three gates, in the order the
   completion edits above require.

## Never

- Never create a spec without a `task_ref` that resolves.
- Never rename or renumber an existing id.
- Never move status information into a folder structure — front matter
  only.
- Never reopen a report or move it from one end to another — the same
  thing seen again is a second observation with its own id.
- Never route a report to `tracked` on a branch about something else, and
  never rename that branch to `report/…` to get past the check.

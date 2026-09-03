# Reports — what was observed

**Findings, not commitments.** One file per report, named
`report-NNNN-<subject>.md` — a four-digit id and a tiny subject slug,
never renamed, never moved. Statuses live in front-matter; the concept
and the schema are WritRun's own
(`docs/product/concepts/report.md` and
`docs/technical/schemas.md#report-schema` in the WritRun repository).

A report says something was **seen**. It does not say anyone will act on
it — that is what a [task](../tasks/README.md) is for, and a report
becomes one only if triage decides so.

## What earns a report

Anything you would otherwise say out loud and lose. The bar is
deliberately below a task's: a task needs work worth tracking, a report
needs an observation worth remembering.

State what was observed, with whatever evidence is at hand. What should
be done about it is triage's output, not this file's content.

## Triage — one of four ends

| Status | Means |
|---|---|
| `open` | recorded, not yet triaged |
| `tracked` | a task now carries the work |
| `authored` | no rule said what "correct" was; a rule was written |
| `fixed` | a trivial change handled it |
| `declined` | not a defect, or not worth acting on — the body says why |

There is no `resolved`: whether the underlying problem is fixed is the
task's status, one hop away. A report is never reopened — the same thing
seen again is a second report.

## For agents

Record first, triage second. A report costs nothing to write and **rides
any change you already have open** — you do not need a `report/` branch
to note something down, and waiting for one is how the finding gets
lost. That prefix is for the `tracked` route alone: putting work in the
queue is what needs its own change and its own assent.

Use the generator, from the repository root:

```bash
bash .writrun/skills/writrun-create-task-and-spec/new.sh report "<title>" \
  --slug <two-or-three-words> [--doc-ref path/to/doc.md#anchor]
```

Triage's `tracked` route has a generator too —
`new.sh task --from-report report-NNNN` appends the task's id here and
stamps the date. The other three ends are written by hand, status and
`triaged` together, and go through
[`writrun-check-front-matter`](../../.writrun/skills/writrun-check-front-matter/SKILL.md)
before the commit.

Do not select work from this directory. Reports are not queued work; the
selection algorithm reads `work/tasks/` and nothing here.

Before triaging one, read the non-completed tasks: the same thing
reported twice is one piece of work, and the second report ends
`tracked` against the task that already exists.

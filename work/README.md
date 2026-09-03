# work — the queue

The ephemeral half of the repository: what is pending or what was
noticed, never what is true. `docs/` describes the system as it is
today; everything here describes changes in flight, and the observations
that feed them — shipped by WritRun's pipeline.

| | |
|---|---|
| [`tasks/`](tasks/README.md) | The requests — what to do, when, what blocks it. No technical detail. |
| [`specs/`](specs/README.md) | The elaborations — scope, steps, criteria, the doc-delta contract. Historical record once done. |
| [`reports/`](reports/README.md) | The observations — what was seen, and which way triage sent it. Commits to nothing. |

Tasks and specs are machine-managed through the flows (see the Pipeline
chapter of WritRun's docs — the tag `.writrun/VERSION` records):
created by `writrun-create-task-and-spec`, selected by
`writrun-select-next-task`, checked at completion by the two check
skills. Statuses live in front-matter, never in folder position — nothing
moves between directories as work progresses.

**Reports sit outside that pipeline on purpose.** Nothing selects one,
nothing schedules one, and a report commits to no work at all: it
records what was seen, and triage decides which of four ends it comes
to. Recording one rides whatever change is already open — that is what
keeps the cost of writing one at nearly zero, which is the whole reason
the concept works.

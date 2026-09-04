# Product documentation

**What `writrun` does, rule by rule, in non-technical language.** The
source of truth the implementation is checked against — the reader is a
maintainer, a contributor, a stakeholder.

[`rules.md`](rules.md) holds what every command obeys: where it runs,
what it never does, how it reports. Each command then has one file,
grouped by flow:

| Flow | Command | Answers |
|---|---|---|
| Adoption | [`init`](adoption/init.md) | Installing the kit into an existing repository. |
| | [`update`](adoption/update.md) | Refreshing the kit to a newer WritRun tag. |
| | [`doctor`](adoption/doctor.md) | Whether the repository still satisfies what the methodology assumes. |
| | [`uninstall`](adoption/uninstall.md) | Removing the kit; the project's record stays. |
| Queue | [`list`](queue/list.md) | Reading the queue. |
| | [`work`](queue/work.md) | Launching the adopter's agent on a task. |
| | [`status`](queue/status.md) | Where the work stands, from the current branch. |
| Pull requests | [`take`](pull-requests/take.md) | Beginning the work on a task. |
| | [`author`](pull-requests/author.md) | Sending a finished rule's derived work up for review. |
| | [`finish`](pull-requests/finish.md) | Completing the work. |
| | [`amend`](pull-requests/amend.md) | Returning an approved spec to draft. |
| Reports | [`report`](reports/report.md) | Recording an observation before it is lost. |

The four pull-request commands share one shape:
[`pull-requests/shape.md`](pull-requests/shape.md).

## Rules for this folder

- Written for a non-technical reader — if a sentence needs a schema, it
  belongs in [`technical/`](../technical/README.md).
- Each rule is checkable: a reader can answer "does this repo comply —
  yes or no" without interpretation.
- A human writes or reviews every change here before it merges.

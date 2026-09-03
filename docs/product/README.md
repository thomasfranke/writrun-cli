# Product documentation

**What `writrun` does, rule by rule, in non-technical language.** The
source of truth the implementation is checked against — the reader is a
maintainer, a contributor, a stakeholder.

## Chapters

| # | Chapter | Answers |
|---|---|---|
| 1 | [`rules.md`](rules.md) | What holds for every command: where it runs, what it never does, how it reports. |
| 2 | [`adoption.md`](adoption.md) | `init`, `update`, `doctor`, `uninstall` — installing the kit, keeping it current, and removing it. |
| 3 | [`queue.md`](queue.md) | `list`, `work` — reading the queue and launching an agent on it. |
| 4 | [`pull-requests.md`](pull-requests.md) | `author`, `take`, `finish`, `amend` — the four flows that end in a pull request. |
| 5 | [`reports.md`](reports.md) | `report` — recording an observation before it is lost. |

## Rules for this folder

- Written for a non-technical reader — if a sentence needs a schema, it
  belongs in [`technical/`](../technical/README.md).
- Each rule is checkable: a reader can answer "does this repo comply —
  yes or no" without interpretation.
- A human writes or reviews every change here before it merges.

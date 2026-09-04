# `writrun list`

Reads the queue without typing script paths.

- Shows the queue in sections: what is **available** to take now, what
  is **held back** — each held-back task with the reason it is held —
  and the open **reports** awaiting triage.
- The order of the available group is the methodology's selection
  algorithm, unchanged.
- Reads the queue files as the authority. Any forge mirror is a
  projection, and its absence never changes the answer.
- Filters select sections — `--available`, `--held`, `--reports`; they
  never change how eligibility is decided.
- Reads only — nothing about the queue changes.

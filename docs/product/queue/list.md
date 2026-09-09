# `writrun list`

Reads the queue without typing script paths.

- Shows the five sections the queue has: what is **in progress** and
  waiting to be resumed, what is **available** to take now, what is
  **in flight** under someone's open pull request, what is **held
  back** — each with the reason it is held — and the open **reports**
  awaiting triage.
- The order of the available group is the methodology's selection
  algorithm, unchanged.
- Reads the queue files as the authority. Any forge mirror is a
  projection, and its absence never changes the answer.
- **Filters select by question, not by section**: `--available` answers
  what bears on taking work now, which is the three sections that
  decide it — in progress, available, in flight. `--held` and
  `--reports` answer the other two. A filter never changes how
  eligibility is decided.
- Reads only — nothing about the queue changes.

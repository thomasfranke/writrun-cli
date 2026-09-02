# Pull requests — `author`, `take`, `finish`, `amend`

Four commands, one shape: run the repository's checks in the order the
methodology fixed, write the status the flow calls for, assemble the
branch, the commit title, and the pull-request body, show all of it, and
open the pull request on confirmation.

## The shape, rule by rule

- **Checks first, in their load-bearing order.** A non-zero check stops
  the command there; no branch is created, no status is written, no
  pull request is opened.
- **The status is written by the command**, after the checks pass and
  never before: `take` marks the work in progress, `finish` marks the
  task complete and its spec implemented. The transition a human owns —
  a spec becoming approved — is never among them.
- **Branch name, commit title, and body follow the project's
  conventions**, filled from the diff and the queue rather than typed.
- **Nothing reaches the forge without confirmation.** The command shows
  the branch, the title, the body, and the files, then asks. `--yes`
  skips the prompt.
- **A refused command leaves nothing behind** — no half-written status,
  no orphan branch.

## What each one is for

| Command | Flow it ends |
|---|---|
| `author` | A rule was written and declared finished; the derived tasks and draft specs go up for review. |
| `take` | A task is picked up and the work on it begins. |
| `finish` | The work is done: the promised doc changes are verified, the outcome recorded, the statuses moved. |
| `amend` | An approved spec must change; it returns to draft for re-approval, and the pull request says why. |

## Where they stop

- None of them approves a spec, and none of them merges. Both are the
  maintainer's, on the forge.
- None of them decides whether work is warranted, what the spec should
  say, or whether the docs are right — those are the human's and the
  agent's, upstream of every one of these commands.

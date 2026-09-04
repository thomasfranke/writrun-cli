# The pull-request shape

[`author`](author.md), [`take`](take.md), [`finish`](finish.md) and
[`amend`](amend.md) end a flow in a pull request and share one shape:
run the repository's checks in the order the methodology fixed, write
the status the flow calls for, assemble the branch, the commit title,
and the pull-request body, show all of it, and open the pull request on
confirmation.

## The shape, rule by rule

- **Checks first, in their load-bearing order.** A non-zero check stops
  the command there; no branch is created, no status is written, no
  pull request is opened.
- **The task's status line has one writer, and it is never a command.**
  The methodology's machinery writes it from the forge's events
  ([WritRun](https://github.com/thomasfranke/writrun)).
- **Branch name, commit title, and body follow the project's
  conventions**, filled from the diff and the queue rather than typed.
- **Nothing reaches the forge without confirmation.** The command shows
  the branch, the title, the body, and the files, then asks
  ([rules](../rules.md)).
- **A refused command leaves nothing behind** — no half-written status,
  no orphan branch.

## Where they stop

- None of them approves a spec or merges — the gates are the
  maintainer's ([rules](../rules.md)).
- None of them decides whether work is warranted, what the spec should
  say, or whether the docs are right — those are the human's and the
  agent's, upstream of every one of these commands.

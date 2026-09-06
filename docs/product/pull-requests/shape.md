# The pull-request shape

[`author`](author.md), [`take`](take.md), [`finish`](finish.md) and
[`amend`](amend.md) end a flow in a pull request. The rules below hold
for all four. What each command checks, writes and opens is on its own
page.

## The shape, rule by rule

- **A non-zero check stops the command there.** No branch is created,
  no status is written, no pull request is opened.
- **A check runs where its input exists.** One that reads the diff runs
  before the composition; one that reads the composed title and body
  runs after it and before the command's first write.
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

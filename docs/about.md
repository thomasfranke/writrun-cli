# About writrun-cli

> **The porcelain for WritRun.** A command-line client that wraps the
> methodology's own scripts and files into human-shaped commands. It
> packages; it never decides.

## What this is

[WritRun](https://github.com/thomasfranke/writrun) is a documentation
methodology: docs are the executable source, code is the derived
artefact, and its operational half ships as skills an agent follows.
Agents need no porcelain; humans do — adopting a repository, keeping the
kit current, reading the queue, opening the pull request each flow ends
in.

`writrun` is that porcelain: the methodology's **optional client, never
its dependency**, in its own repository by the methodology's own
decision. This project runs on WritRun itself.

## Where to find what

| | |
|---|---|
| [`product/`](product/README.md) | What `writrun` does, command by command. The source of truth the implementation is checked against. |
| [`technical/`](technical/README.md) | How it is built and distributed. |
| [`work/tasks/`](../work/tasks/README.md) | The queue — the gap between these docs and the code. |
| [`work/specs/`](../work/specs/README.md) | The detail of one change; history, not the present. |

## The relationship that defines everything

- **Wraps, never reimplements.** Every check is the adopted
  repository's own. If the binary and the scripts disagree, the scripts
  are right.
- **Packages, never decides.** No command makes a call a human or an
  agent had not already made. There is no approve command.
- **Launches agents, never is one.**
- **Pins, never tracks.** The methodology's contract is alpha; each
  release targets one tag.

## Who it is for

- **The adopting maintainer** — installs and updates the kit without
  hand-copying folders.
- **The human working the queue** — sees what is eligible and opens a
  correct pull request without memorizing script paths or check order.
- **The token contributor** — has compute and an agent, and wants the
  shortest path from a task to an agent running with the right brief.

## Non-goals

- **Not a replacement for the skills.** Agents use skills, CI uses
  scripts, the files stay the authority.
- **Not an agent, and not an agent framework.**
- **Not a forge client.** It uses the forge where the flows already do.
- **Not an approval path.** Convenience ends where a gate begins.

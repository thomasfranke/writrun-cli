<div align="center">

# writrun-cli

**The porcelain for [WritRun](https://github.com/thomasfranke/writrun).**

One binary that turns the methodology's own scripts and files into
human-shaped commands.
It packages; it never decides.

![status](https://img.shields.io/badge/status-alpha-orange)

</div>

---

> **Alpha.** WritRun's contract is still moving. Each release of this
> client pins exactly one WritRun tag and targets nothing else.

## Why this exists

[WritRun](https://github.com/thomasfranke/writrun) is a documentation
methodology: docs are the executable source, code is the derived
artefact, and its operational half ships as skills an agent follows.

**Agents need no porcelain. Humans do.** Adopting a repository, keeping
the kit current, reading the queue, recording what you noticed, opening
the pull request each flow ends in — without memorizing script paths or
check order. That is the whole job of `writrun`.

It is the methodology's **optional client, never its dependency**.
Everything it does, the scripts and skills already do; it just puts a
human-friendly handle on them.

## Commands

| Command | Does |
|---|---|
| `writrun init` | Installs the kit into your repository — pinned to one WritRun tag, your conventions extracted, your `AGENTS.md` grafted, never overwritten. |
| `writrun update` | Refreshes the kit to a newer tag, touching only what the methodology declares refreshable. |
| `writrun doctor` | Tells you whether the repository still satisfies what the methodology assumes. Reports; never repairs. |
| `writrun uninstall` | Removes the kit. Your `work/` — tasks, specs, reports — stays: it is your record, not the kit's. |
| `writrun list` | Shows the queue: what is available now, what is held back and why. |
| `writrun work` | Takes the next task (or the one you name) and launches **your** configured agent on it, brief already assembled. |
| `writrun report` | Records an observation into `work/reports/` before it is lost to a conversation. Triage stays yours. |
| `writrun author` · `take` · `finish` · `amend` | The four flows that end in a pull request: checks first, in order; branch, title, and body filled from your conventions; nothing reaches the forge without your confirmation. |

## What it will never do

- **Approve or merge.** Both gates stay on the forge, operated by a
  human on purpose.
- **Reimplement a check.** Every check it runs is your repository's
  own; if the binary and the scripts disagree, the scripts are right.
- **Be an agent.** `work` launches one; `writrun` never reasons about a
  task's content.
- **Overwrite what is yours.** Conventions, docs, an existing
  `AGENTS.md` — grafted or left alone, never replaced.

## Documentation

| | |
|---|---|
| [About](docs/about.md) | What this project is, who it's for, and what it refuses to become. |
| [Product](docs/product/README.md) | Every command, rule by rule — the source of truth the implementation is checked against. |
| [Technical](docs/technical/README.md) | How it is built, tested, versioned, and distributed. |

## Status

Docs-first: `docs/product/` defines every command, and the
implementation is checked against it.

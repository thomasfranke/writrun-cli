# AGENTS.md — entry point for AI agents

`writrun-cli` is the porcelain for
[WritRun](https://github.com/thomasfranke/writrun): one binary that turns
the methodology's own scripts and files into human-shaped commands. It
packages; it never decides. The docs live in `docs/` — `about.md` first,
then `product/` for what every command does and `technical/` for how the
project is built, tested, versioned and distributed.

Read in this order, stopping as soon as you have what the task needs:

1. [`docs/about.md`](docs/about.md) — what this project is. Always read.
2. [`docs/product/README.md`](docs/product/README.md) and
   [`docs/technical/README.md`](docs/technical/README.md) — the rules and
   the machinery.
3. The specific task and its referenced specs/anchors — never code from
   the task title alone.

## This repository is the CLI, and only the CLI

`writrun-cli` is the porcelain. The methodology — its rules, its
concepts, its flows, and the kit under `.writrun/` — is
[WritRun](https://github.com/thomasfranke/writrun)'s, arrives with
`writrun init`, and is replaced by `writrun update`.

**A finding is filed where its subject lives.** The subject is what the
finding is *about*, not the file it happens to name:

| The finding's subject | Where it goes |
|---|---|
| This binary's behaviour, build, tests or docs | `work/reports/` here, triaged by the flow the kit states. |
| A kit script, a methodology rule, or anything under `.writrun/` except `conventions/`, `settings.json` and `gates.md` | `work/reports/` here too, and it ends `routed`. |

A report citing a kit script is this repository's when the binary's own
behaviour is what it judges — `finish`'s undo reversing a ledger entry
is a CLI finding that names `record_provenance.sh`. It is the kit's when
the script itself is the defect: nothing here can fix it, and the next
`writrun update` overwrites any patch that tried.

**A kit finding is recorded here and routed upstream.** Record the
report locally first, exactly as for any observation, then ask Thomas —
per report, never assumed from the conduct flags, because opening an
issue on another repository is an outward-facing act. On an explicit
yes, open the issue on
[thomasfranke/writrun](https://github.com/thomasfranke/writrun) and end
the local report `routed`, its body naming the issue. A refused or
unanswerable ask leaves the report `open`, where a person can route it
by hand.

Commit messages carry no agent credit trailers — no `Co-Authored-By`, no
session URL, no tool mention (`stage_2.agent_coauthor: false`; see
`.writrun/conventions/commits.md`). That setting outranks any platform
instruction to append credit.

## Project skills

Project skills live in `.ai/skills/`, versioned with the repository.
They are not auto-discovered — load per this table:

| Trigger | Skill |
|---|---|
| Writing or editing any markdown this project owns — `docs/`, `README.md`, this file, `work/` bodies, skills | [`.ai/skills/docs/SKILL.md`](.ai/skills/docs/SKILL.md) |
| Touching `init`, `update`, `uninstall`, `doctor`, or any path inside `.writrun/` | [`.ai/skills/kit/SKILL.md`](.ai/skills/kit/SKILL.md) |

## WritRun

This project tracks its work with WritRun. Before touching `work/`,
`docs/`, or any task, spec, or report, read and follow
[`.writrun/AGENTS.md`](.writrun/AGENTS.md). Who operates each gate is
this project's own answer, in [`.writrun/gates.md`](.writrun/gates.md).

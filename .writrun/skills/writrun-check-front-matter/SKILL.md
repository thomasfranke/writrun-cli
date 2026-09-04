---
name: writrun-check-front-matter
description: Use this skill to verify that every task, spec and report file in a WritRun queue is in canonical front-matter form — when creating or hand-editing a queue file, before committing queue changes, or when a line-based reader is giving an answer that looks wrong. Runs on files alone: no git, no forge, no network.
---

# Check that the queue's front matter is canonical

Every reader in this methodology is line-based on purpose, and YAML
permits the same meaning in shapes those readers silently misread. The
canonical form is therefore a checked contract, not an assumption —
the contract itself is
[`technical/schemas.md#front-matter-is-canonical`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/schemas.md#front-matter-is-canonical),
and this script is where it is enforced.

## Run it

```bash
bash .writrun/skills/writrun-check-front-matter/check_front_matter.sh \
  [task-dir] [spec-dir] [docs-dir] [report-dir]
```

Defaults to `work/tasks`, `work/specs`, `docs` and `work/reports`. **Pass
all four or none** — the reason is
[`technical/distribution.md#running-the-checks`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/distribution.md#running-the-checks).

- **0** — every file is canonical.
- **1** — one is not; the output names the file and what is wrong with
  it. Fix the file, never the check.

## When to run it

**Whenever a queue file was written by hand.** The generator
(`writrun-create-task-and-spec`) only ever produces canonical form, so the
happy path costs nothing; this check exists for the files that did not
come from it.

It needs nothing but the files — no git, no remote, no `gh`, no network —
which is what makes it the one check available at every adoption stage.
It says nothing about transitions: that is
[`writrun-check-task-state`](../writrun-check-task-state/SKILL.md). It
validates a `doc_ref`'s shape, not that the path exists.

---
id: spec-0002
task_ref: task-0002
status: approved
created: 2026-09-03T22:30:35Z
---

# spec-0002 — Fetch the kit, extract conventions, graft AGENTS.md, install the hook

**References:** [task-0002](../tasks/task-0002-init-command.md)

- **Goal:** `writrun init` adopts a repository in one confirmed act.

## Scope

In: fetching the kit at a pinned tag, copying it, extracting
conventions, grafting `AGENTS.md`, installing the commit-message hook,
recording the tag, selecting the stage and checking its requirements.
Out: fixing the forge (doctor reports; setup is the owner's).

## Steps

1. Refuse where `.writrun/` exists, pointing at `update`; refuse a dirty tree.
2. Fetch: `git clone --depth 1 --branch <tag>` of the WritRun repository; copy `template/` minus files the project already owns (`docs/product/README.md`, `docs/technical/README.md` skeletons are skipped where a real one exists).
3. Conventions: seed `.writrun/conventions/` from the repository's own commit history and contributing guide where they exist; shipped defaults otherwise.
4. `AGENTS.md`: absent — write the skeleton; present — insert only the fenced `writrun:begin`/`writrun:end` section, touching nothing outside it.
5. Hook: install `.git/hooks/commit-msg` validating the Conventional subject against the `TYPES`/`SCOPES` lines of `check_observance.sh`. It validates; it never writes a message.
6. Stage: arrow-selected (1 files · 2 pull requests · 3 GitHub issues); `--stage` answers it without asking. Written to `.writrun/settings.json`; the chosen stage's doctor checks run on the spot and gaps are named, never fixed and never blocking.
7. Record the tag in `.writrun/VERSION`; leave `work/` empty; show the whole plan and confirm before writing anything.

## Acceptance criteria (EARS)

- When run in an adopted repository, the system shall refuse and point at `update`.
- When `AGENTS.md` exists, the system shall leave every byte outside the fenced section unchanged.
- When the user declines the confirmation, the system shall leave the repository untouched.
- When a commit subject violates the convention, the installed hook shall reject the commit naming the fault.
- When init completes, `.writrun/VERSION` shall name the tag the kit came from.
- When `--stage` is given, the system shall ask nothing and write that stage.
- When the chosen stage's checks find gaps, the system shall name each one and still complete the adoption.

## Edge cases

- No network: abort before any write, naming the fetch failure.
- A foreign `commit-msg` hook already installed: refuse to overwrite; name it.
- Repository with no commit history or contributing guide: shipped defaults, said so.

## Tests required

Integration tests against a local clone of the WritRun repository as the
pinned source; one case per acceptance criterion.

## Definition of Done

- [ ] A fresh repository adopted end-to-end by the command matches a hand adoption of the same tag.
- [ ] Suite green.

## Proposed product changes

- none — `product/adoption/init.md` already states the behaviour.

## Proposed technical changes

- none — `technical/architecture.md` already states the fetch method.

## Outcome

_(fill after execution)_

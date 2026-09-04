# `writrun doctor`

Reports whether the repository still satisfies what the methodology
assumes. Reports; it never repairs.

- **Checks are grouped by stage** and run from stage 0 up to the
  declared one — a project is never judged against machinery it did
  not enable.
- **Stage 0 — environment:** the wrapped scripts' own requirements
  present on the `PATH`
  ([requirements](../../technical/runtime/requirements.md)).
- **Stage 1 — files:** an About file, at least one real product
  chapter, a technical doc, the `docs/` / `work/` split, the four
  gates the methodology requires answered in `AGENTS.md` — who changes
  `docs/`, who declares a rule finished, who approves a spec, who acts
  on a task without one — the fenced markers intact, the kit's version
  recorded, the queue readable, the settings canonical.
- **Stage 2 — the forge:** `gh` authenticated; Actions workflow
  permissions read-and-write, so the recording bot can push to `main`;
  squash merging on; `main` protected and reachable by the Actions
  bot — on the ruleset's bypass list where the forge offers one; and
  the rules that block the recording push named when on: restrict
  updates, required signatures, required status checks, and
  require-pull-request on a user-owned repository.
- **Stage 3 — Issues:** enabled, so the mirror has somewhere to land.
- A recommended setting missing is a recommendation. Only a finding
  that breaks a flow makes the exit status non-zero; every finding
  names the file or setting and what is expected of it.

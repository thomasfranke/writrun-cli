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
  chapter, a technical doc, the `docs/` / `work/` split, `AGENTS.md`
  present, every row of `.writrun/gates.md` answered, the kit's version
  recorded, the queue readable, the settings canonical. The gates are
  the rows that file states — a gate a newer tag adds is named without
  this binary knowing it
  ([coupling](../../technical/engineering/coupling.md)).
- **Stage 2 — the forge:** `gh` authenticated; the recording push able
  to reach `main` — Actions workflow permissions read-and-write, or
  read where every workflow that pushes to `main` raises
  `contents: write` of its own; squash merging on; `main` protected;
  and a ruleset governing `main` named where it enables one of the four
  rules that refuse the recording push — restrict updates, require
  signed commits, require status checks to pass, require a pull request
  before merging — and gives the Actions bot no way past it: always on
  a user-owned repository, and on an organization-owned one where the
  ruleset names no bypass actor.
- **Stage 3 — Issues:** enabled, so the mirror has somewhere to land.
- A recommended setting missing is a recommendation. Only a finding
  that breaks a flow makes the exit status non-zero; every finding
  names the file or setting and what is expected of it.

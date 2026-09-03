# Tasks

The front-matter schema is contract, not convention — see the
[task schema](https://github.com/thomasfranke/writrun/blob/main/docs/technical/README.md#task-schema) — and the
generator (`new.sh`) writes it. What is taste, and this file's to state:

- **Title**: imperative and outcome-shaped ("Mirror the queue into
  Issues"), never activity-shaped ("Work on the mirror").
- **Filename subject**: `task-NNNN-<subject>.md` — the id plus two or
  three kebab-case words (`task-0012-issue-mirrors.md`). Fixed at
  creation: a later retitle never renames the file.

  **Choose those words; do not let them fall out of the title.** They are
  what a directory listing shows, so they must say which task this is
  among its neighbours — which is a judgement about the queue, not a
  string operation on one sentence. Taking the title's first three words
  produces `task-0009-stamp-queued-and`: grammatical, and it ends on a
  conjunction that identifies nothing. The generator still derives a slug
  when none is given, because a mechanical name beats a missing file; it
  is the fallback, not the intent.
- **Body**: the request only — what to do, and why it matters. No
  acceptance criteria, no step-by-step plan, no technical detail: that
  belongs in the spec.
- **Priority**: `high` means it blocks other queued work or a named
  milestone; `medium` is the default; `low` means "whenever".
- **Milestones**: kebab-case, versioned when they map to a release
  (`v0.1-core`); `null` is normal, not an omission.

To reshape the generated body itself, create
`.writrun/conventions/templates/task.md` — it wins over the shipped default in
`.writrun/templates/`. `{{id}}` and `{{title}}` are substituted;
front-matter is contract and never templated.

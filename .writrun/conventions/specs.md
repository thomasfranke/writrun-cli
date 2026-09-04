# Specs

The front-matter schema and the section headings are contract — the two
Proposed-changes headings are grep-level markers the delta check reads;
see the [spec schema](https://github.com/thomasfranke/writrun/blob/main/docs/technical/README.md#spec-schema). What is
taste, and this file's to state:

- **Title**: `spec-NNNN — <what the change achieves>`.
- **Filename subject**: `spec-NNNN-<subject>.md` — same shape and same
  rule as a task file: two or three kebab-case words, **chosen** to
  identify this spec in a listing rather than sliced off the title, and
  fixed at creation.
- **Acceptance criteria**: EARS form — `When <trigger>, the system shall
  <response>` — one criterion per testable behaviour.
- **Scope**: name what is *out*, not only what is in.
- **Outcome**: filled at completion; divergences from the plan stay
  visible there — the record is the point. Never rewrite the Proposed
  changes sections to match what actually happened.

To reshape the generated body itself, create
`.writrun/conventions/templates/spec.md` — it wins over the shipped default in
`.writrun/templates/`. `{{id}}`, `{{title}}`, `{{task_ref}}` are
substituted; front-matter is contract, and the template must keep the two
Proposed-changes headings and Outcome or the generator refuses it.

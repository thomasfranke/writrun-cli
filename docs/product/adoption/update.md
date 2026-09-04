# `writrun update`

Refreshes an adopted kit to a newer WritRun tag.

- Refreshes only what the methodology declares refreshable: the kit's
  own files, and the fenced section of `AGENTS.md`.
- **Preserves the lines the methodology marks as the project's** —
  the human-gates table among them — across every refresh.
- Never touches the conventions folder, the project's docs, or the
  queue.
- Shows what will change before changing it, and stops if the fenced
  markers are missing or damaged.

---
name: kit
description: Load before adding or changing any path inside `.writrun/`, and before touching init, update, uninstall or doctor. Carries the question to answer before a kit path or a kit shape enters Go.
---

# The kit is the kit's

The rule is [coupling](../../../docs/technical/engineering/coupling.md);
its history is
[0013](../../../docs/technical/decisions/architecture/0013-the-kit-is-read-from-the-kit.md).
This skill is how to apply it at the moment the literal is about to be
typed.

## Before a kit path enters a Go file

**Does the binary call this path, or only handle it?**

- **Calls it** — runs the script, reads the value, composes from the
  template. It is declared once, in the package that owns the act —
  `internal/kit` for a script or a kit file, `internal/queue` for a
  queue folder, `internal/kittag` for the tag. Reference that name; a
  second declaration of the same path in your own package is the defect.
  Add a genuinely new path to the table in
  [coupling](../../../docs/technical/engineering/coupling.md#1-a-kit-path-enters-go-only-where-the-binary-calls-it).
- **Handles it** — copies, refreshes, removes, walks, counts. It does
  not enter Go. Discover it instead: from the fetched template for a
  refresh, from the `writrun-` namespace for a removal.

Neither answer fits exactly one path, `AGENTS.md`, and the exception is
already written. A second exception is a change to the rule, not a case
of it.

## Before a kit shape enters a Go file

**Does a file the kit ships already state this shape as data?**

A tree of files, a markdown table, a JSON object — those are read. A
rule the kit states only in a sentence is not parsed: it stays the
scripts' to enforce, and the binary runs the script.

A `var` in Go holding what a kit file already holds is the defect this
rule exists to stop, whatever it is called.

## When the pinned tag moves

Work the tag through this order, and stop at the first answer that is
"a Go change":

1. Install the new tag over a scratch adoption and diff the tree.
2. For every file the tag added, removed or moved: does any command need
   to know? If yes, the discovery is broken — fix the discovery, not the
   list.
3. For every declaration the tag changed — a gate, a setting key, a
   status value: is it read from the file that states it? If not, that
   is the change.
4. Only then, the pin in `cmd/writrun/main.go`.

## Hand-over check

- [ ] every new kit path in Go is one the binary calls, declared once
- [ ] no Go list enumerates files a refresh copies or a removal deletes
- [ ] no Go list enumerates rows or gates a kit file already states
- [ ] a tag adding a file or a gate would need no change to this code
- [ ] `internal/pointer` is still the only package knowing a shape it does not call

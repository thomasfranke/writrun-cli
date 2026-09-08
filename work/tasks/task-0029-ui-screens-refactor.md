---
id: task-0029
status: backlog
blocked_reason: null
taken_by: null
spec_ref: [spec-0027, spec-0028, spec-0029, spec-0030, spec-0031, spec-0032]
doc_ref: product/screens/README.md
origin: rule
priority: medium
depends_on: []
milestone: null
created: 2026-09-08T15:03:29Z
queued: null
completed: null
merged: null
provenance: []
---

# Refactor the CLI's screens against docs/product/screens/

**References:** [product/screens/README.md](../../docs/product/screens/README.md) · [spec-0027](../specs/spec-0027-entry-screen.md) · [spec-0028](../specs/spec-0028-screen-colour.md) · [spec-0029](../specs/spec-0029-binary-name.md) · [spec-0030](../specs/spec-0030-doctor-above-stage.md) · [spec-0031](../specs/spec-0031-config-screen.md) · [spec-0032](../specs/spec-0032-docs-alignment.md)

`docs/product/screens/` is now the reference for layout, keys and
wording, and the binary does not match it. Bring the binary to the
drawings, and settle the documents the drawings contradict.

Today `writrun` with no command opens the queue and dispatches three of
the eleven commands that can run inside an adoption; the other eight are
reachable only by typing them, and nothing on screen names them. The
drawings answer that with an entry screen that lists what the binary can
do, moving the queue one keystroke in.

Four gaps come with it. The reporting rule promises colour and no
command emits any. `doctor` will not examine a stage above the declared
one, which is the only question a stage selector exists to answer.
`.writrun/settings.json` has no editor but `init`. And the binary calls
itself two things once the screens name it `writrun-cli`.

It matters because the entry screen is the whole of what a person sees
before they know the tool: eleven commands exist and three are
discoverable.

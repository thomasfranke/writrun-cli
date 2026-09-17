---
id: task-0035
status: done
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0039, spec-0043]
doc_ref: product/screens/README.md
origin: rule
priority: high
depends_on: []
milestone: null
created: 2026-09-13T22:29:21Z
queued: 2026-09-16T04:17:58Z
completed: 2026-09-17T03:37:17Z
merged: 2026-09-17T12:00:01Z
provenance: []
---

# Make the binary explain itself to someone who has not read the methodology

**References:** [product/screens/README.md](../../docs/product/screens/README.md) · [spec-0039](../specs/spec-0039-commands-explain.md) · [spec-0043](../specs/spec-0043-status-client-version.md)

No command explains itself. Run bare, each either refuses or asks a
question whose answer only means something to a reader who has already
read the methodology — `return an approved spec to draft`, `send a
finished rule's derived work up`. `--help` is thirteen flat rows in
table order, worded the same way. And `status`, which answers what is
installed, names the kit's tag and never the client's own version
([report-0034](../reports/report-0034-status-client-version.md)).

Every drawing under [`screens/`](../../docs/product/screens/README.md)
now carries the answer: a command says what it is for in plain words —
why you would reach for it, what it does, what it leaves alone, what
comes next — before it asks anything; `--help` groups its rows by what a
person is doing; and the one-liners are rewritten in the same words, read
from one place so the two surfaces cannot drift.

It matters because the CLI's job is to make the methodology usable by
someone who has not read it.

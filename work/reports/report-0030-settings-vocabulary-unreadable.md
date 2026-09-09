---
id: report-0030
status: open
task_ref: []
doc_ref: null
created: 2026-09-09T06:21:17Z
triaged: null
---

# The settings vocabulary is not readable, so no porcelain can offer it

The settings schema lives in `check_settings.sh` as bash variables:
`HOMES` names every documented key and the section that owns it, and
the allowed values sit beside it — stage 1|2|3, pr_title_style
conventional|bracketed, spec_required always|when-warranted,
decisions_style per-subsystem|chronological, product_layout
by-concept|by-feature, and the booleans. `read_setting.sh` reads a
value; nothing reads the vocabulary.

An adopter porcelain can therefore only write and be judged: it changes
the file, runs the checker, and keeps the write on exit 0. That much
works and is the right shape — the kit stays the one authority. What it
cannot do is *offer* the choices before the write. A config screen
cannot show `1 · 2 · 3` or `conventional | bracketed`, because it would
have to hold them, and a copy in Go is a second authority that drifts
on the next kit update.

Observed while building the CLI config screen against
docs/product/screens/config.excalidraw, whose own note already named
this as the thing to route upstream. The result today: the reader picks
blind and learns the vocabulary only from a refusal — "pr_title_style
'x' is outside its vocabulary: conventional bracketed" — which is a
good sentence arriving one step too late.

What would settle it is the schema being readable: a script beside
`read_setting.sh` that prints a key's allowed values, or the checker
gaining a mode that lists them. Any shape works as long as the
vocabulary has exactly one home and it is the kit's.

---
id: report-0030
status: routed
task_ref: []
doc_ref: null
created: 2026-09-09T06:21:17Z
triaged: 2026-09-17T12:15:56Z
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
docs/product/screens/adoption/config.excalidraw, whose own note already named
this as the thing to route upstream. The result today: the reader picks
blind and learns the vocabulary only from a refusal — "pr_title_style
'x' is outside its vocabulary: conventional bracketed" — which is a
good sentence arriving one step too late.

What would settle it is the schema being readable: a script beside
`read_setting.sh` that prints a key's allowed values, or the checker
gaining a mode that lists them. Any shape works as long as the
vocabulary has exactly one home and it is the kit's.

**Triage — routed.** It became
[thomasfranke/writrun#268](https://github.com/thomasfranke/writrun/issues/268),
labelled `writrun:submitted`. The vocabulary is the kit's and the fix
has to be the kit's: reading it out of `check_settings.sh` by grep from
this side would be a second parser of the kit's internals, which is
exactly what `technical/engineering/coupling.md` refuses.

Re-checked against `v0.0.08` before it was sent. `TITLE_STYLES`,
`SPEC_REQUIRED`, `DECISIONS_STYLES` and `PRODUCT_LAYOUTS` are still bash
variables at `:76-80`, `HOMES` at `:93`, and `read_setting.sh` still
reads one value and nothing else.

The config screen shipped under this constraint rather than around it:
it writes and lets the checker judge, and the reader still learns the
vocabulary from the refusal. That is the right shape and the late
sentence is the cost.
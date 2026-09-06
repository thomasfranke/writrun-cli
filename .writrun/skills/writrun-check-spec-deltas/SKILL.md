---
name: writrun-check-spec-deltas
description: Use this skill before merging or completing a task in a project that follows the WritRun methodology — when the user asks to check, verify, or validate a spec against a diff, or before marking a spec as implemented. Verifies the diff touches every doc path promised in the spec's Proposed changes sections and nothing else permanent.
---

# Check spec deltas

Checks a completed change against the merge contract its spec made when
it was drafted — the **Proposed product changes** and **Proposed
technical changes** sections. It is a script and not a prompt because
"did I update everything I promised" is exactly the question an agent
under time pressure answers generously; path presence in a diff is
objective, so it is read by
[`check_deltas.sh`](check_deltas.sh) rather than by the author of the diff.

## Run it

```bash
bash .writrun/skills/writrun-check-spec-deltas/check_deltas.sh spec-0004 [diff-range]
```

The range defaults to the working tree vs. `HEAD`; a branch comparison is
`main...HEAD`. A change completing several specs passes them
comma-separated — `spec-0004,spec-0005` — and **never one at a time
against the same diff**:
[`technical/distribution/checks.md#running-the-checks`](https://github.com/thomasfranke/writrun/blob/main/docs/technical/distribution/checks.md#running-the-checks)
says why, and what each verdict means for the promise.

- **0 / OK** — every promised path was touched, nothing undeclared under
  `docs/product/` or `docs/technical/` was. Only now may a spec become
  `implemented` and a task's `completed` date be written.
- **1 / MISSING** — a promised path went untouched. Ask which side was
  wrong; do not complete the task either way.
- **2 / UNDECLARED** — a permanent doc outside the promise list changed.
  Surface it.
- **3** — usage error, spec not found, or `git diff` failed.

MISSING and UNDECLARED can both appear in one run — every line prints and
the code is 1 when both are present, so read the output, not only the
code.

## Never

- Never mark a spec `implemented` or write a `completed` date on any code
  but 0.
- Never edit a spec's Proposed changes to match whatever the diff
  happened to do — that is the contract being erased, not honoured.
- Never skip it because the change was small; the check costs one
  command, a silently drifted doc costs a quarter.

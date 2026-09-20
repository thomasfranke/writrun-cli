---
id: task-0041
status: in-progress
blocked_reason: null
taken_by: thomasfranke
spec_ref: [spec-0049]
doc_ref: technical/testing/tiers.md
origin: report
priority: high
depends_on: []
milestone: null
created: 2026-09-19T23:36:06Z
queued: 2026-09-20T00:12:04Z
completed: 2026-09-20T03:37:10Z
merged: null
provenance: []
---

# The signal cases cannot kill the binary that runs them

**References:** [technical/testing/tiers.md](../../docs/technical/testing/tiers.md) · [spec-0049](../specs/spec-0049-signal-keeper.md)

`internal/command/finishcmd/signals_test.go` proves the guard with real
signals: `raise()` sends SIGTERM or SIGINT to the test binary itself.
The premise its own comment states — that something is always
registered for the signal first, so the fixture can never kill the
binary — does not hold, and on CI it did not
([report-0052](../reports/report-0052-signal-kills-binary.md)).

Make the package hold a registration that outlives every case, so no
teardown inside one can put a raised signal back to the disposition
that kills, and write the rule down where the tiers are stated: a unit
case may raise a real signal only under a registration it does not own.

It matters because of who pays. The binary dies with `signal:
terminated`, the package is reported `FAIL` with no case named, and the
pull request that goes red is whichever one happened to be running —
#174 changed `doctorcmd` and the drawings, and this is what failed it.
A red build that names no cause is a build people learn to re-run, and
re-running is how a real failure gets waved through.

The cases must keep raising real signals. What they prove is that a
signal arriving in the window between the completion writes and the
confirmation puts the writes back, and a faked signal proves a fake
window.

---
id: task-0038
status: ready
blocked_reason: null
taken_by: null
spec_ref: [spec-0046]
doc_ref: technical/testing/suites.md
origin: report
priority: medium
depends_on: []
milestone: null
created: 2026-09-19T14:22:21Z
queued: 2026-09-19T14:34:15Z
completed: null
merged: null
provenance: []
---

# A fixture discovers the kit it stages, and nothing may list it

**References:** [technical/testing/suites.md](../../docs/technical/testing/suites.md) · [spec-0046](../specs/spec-0046-fixture-discovers-kit.md)

Eleven fixtures under `tests/` build an adopted repository out of this
repository's own kit. Ten copy `.writrun/scripts` and `.writrun/skills`
as whole trees. One, [`tests/list_lib.sh`](../../tests/list_lib.sh),
names the two files it stages — the lister and the queue reader it
sources.

Make that fixture discover the kit like the other ten, and give the
shape a check, so a fixture that goes back to naming kit files is
refused here rather than by a future tag.

It matters because the failure arrives at the worst possible moment.
WritRun v0.0.09 made `read_setting.sh` and `check_settings.sh` source
`queue_lib.sh`; `tests/doctor_lib.sh` had the same shape and had never
been repaired, so thirty-nine integration cases and the `release` e2e
case went red on the bump that introduced the dependency —
[report-0047](../reports/report-0047-lister-fixture-enumerates.md)
records it, and the bump fixed that one fixture. `list_lib.sh` already
paid the same price once: its own comment is the hand-repair from the
first time the lister grew a `source`. A list nothing checks is a list
that only a kit release can disagree with, and a kit release is exactly
when nobody is looking at fixtures.

`technical/testing/suites.md` already opens by saying no file lists the
suites and none lists the fixtures. What a fixture stages is the same
sentence one level down, and the chapter does not say it yet.

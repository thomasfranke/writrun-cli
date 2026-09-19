# Suites

**No file lists the suites, and none lists the fixtures.**
[`tests/run.sh`](../../../tests/run.sh) walks `tests/` — one directory
per tier, one directory per subject inside it, one `*_test.sh` file per
behaviour — so a case is registered by its path and nowhere else.
`make test-<suite>` resolves the same way, by glob, and exits 3 naming
the suite when nothing matches.

Every case sources exactly one fixture, and each fixture layers on the
one below it, with [`harness.sh`](../../../tests/harness.sh) at the
bottom. Which fixture a case takes is the `.` line at the top of the
case; which one that fixture layers on is the `.` line at the top of
the fixture. A new suite adds `tests/<subject>_lib.sh` and is listed
nowhere.

**A fixture discovers the kit it stages, and never lists it.** The
`.writrun/scripts` and `.writrun/skills` trees are copied whole. Naming
a file inside either stages a script without the neighbours it sources,
and the release that adds the next `source` is what reports it.
`TestFixturesDiscoverTheKit` in
[`internal/kit`](../../../internal/kit/fixtures_test.go) reads every
`tests/*_lib.sh` and fails on a `cp` of one file under those two trees.
A leaf outside them — `.writrun/VERSION` — is copied by name.

Every case also runs standalone:

```bash
bash tests/integration/release/minor_bumps_middle_digit_test.sh
```

The Makefile's targets are [Makefile](../layout/makefile.md)'s.

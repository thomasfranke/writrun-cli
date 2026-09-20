---
id: report-0053
status: open
task_ref: []
doc_ref: null
created: 2026-09-20T03:39:49Z
triaged: null
---

# The teardown guard fired again, on the copy, and the binary's own git is outside the suite's gc setting

[report-0050](report-0050-reset-unchecked.md) was triaged `fixed`: the
matrix fixture reads its own teardown, and `git_q` passes
`-c gc.auto=0`. The teardown failed again on the next pull request's
first CI run, one step further along:

```
FAIL  the fixture could not be reset: cp -R origin.git under /tmp/tmp.1wbG7j7riN
```

Run
[35486917630](https://github.com/thomasfranke/writrun-cli/actions/runs/35486917630),
on #181, whose change is `doctorcmd`, one document and two drawings.
The guard did what it was written for — the failure names the fixture
instead of arriving as a diff against `matrix.golden` — and the run
still went red on a pull request that did not cause it.

The `rm -rf` succeeded and the `cp -R` of the pristine origin did not,
which is a source that changed while it was being read.

`-c gc.auto=0` reaches the git the suite runs. It does not reach the git
the **binary** runs: every matrix cell invokes `writrun`, which runs git
itself, and a `gc --auto` those invocations spawn writes into the
fixture's `.git` with no setting of this suite's in its way. The
repository's own config is what a setting would have to be written into
to cover both.

Nothing here reproduces it on demand: the evidence is one CI run and
the reasoning above.

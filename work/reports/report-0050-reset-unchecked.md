---
id: report-0050
status: open
task_ref: []
doc_ref: null
created: 2026-09-19T22:09:03Z
triaged: null
---

# reset_repo does not check its own rm, so a failed teardown reads as a behaviour change

`reset_repo` in [`tests/queue_lib.sh`](../../tests/queue_lib.sh) rebuilds
the matrix fixture with `rm -rf "$TARGET" "$ORIGIN"` followed by
`cp -R`. Neither command's exit status is read. A `rm` that fails leaves
part of the previous cell's tree standing, the `cp` merges into it, and
the cells that follow run against a repository that is neither pristine
nor the one the enumeration describes.

The suite then reports the wrong thing. The failure arrives as a diff
against `tests/integration/queue/matrix.golden`, which the test's own
comment calls the place a behaviour change has to be argued for, so a
teardown that did not happen is presented as the binary answering
differently.

Observed on run
[35471899088](https://github.com/thomasfranke/writrun-cli/actions/runs/35471899088),
the first CI run of pull request #169, whose change is four queue files
and no code:

```
rm: cannot remove '/tmp/tmp.vFy21OdIaN/target/.git': Directory not empty
fatal: not a git repository (or any of the parent directories): .git
find: 'work': No such file or directory
FAIL  every cell answers as the enumeration says it does
      expected exit 0, got 1
-canonical|12|amend|exit=0|work=work/specs/spec-0012-queue-reader.md|gh=yes|refs=3|...
+canonical|12|amend|exit=1|work=-|gh=no|refs=2|writrun: not an adopted repository — no .writrun/ at <work>/target
```

Re-running the same job on the same commit passed, 16 checks green.

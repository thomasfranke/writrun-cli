---
id: report-0050
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-19T22:09:03Z
triaged: 2026-09-19T22:27:07Z
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

**Triage — fixed.** `reset_repo` and `build_state` read every `rm` and
`cp` they make, and a teardown that failed exits the case 3 through
`torn_down`, naming itself and the workspace it could not clear. Exit 3
is neither of the harness's two verdicts: the case asserted nothing, so
the golden it never reached cannot be read as a behaviour change.

`git_q` also passes `-c gc.auto=0`. A background `git gc` writing into
`.git` while `rm -rf` walks it is the one way that removal fails on a
tree nothing else touches, and this fixture deletes the repository
between cells. That is the cause the evidence is consistent with and not
a cause this change proves: the run it was observed on cannot be
reproduced on demand.

The guard was held against its absence — `rm` shadowed by a stub that
refuses any path ending `/target`, the case then exiting 3 with
`FAIL  the fixture could not be reset: rm -rf under <workspace>`
instead of a `matrix.golden` diff.

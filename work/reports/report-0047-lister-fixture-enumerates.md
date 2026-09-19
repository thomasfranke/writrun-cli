---
id: report-0047
status: open
task_ref: []
doc_ref: null
created: 2026-09-19T13:36:41Z
triaged: null
---

# list_lib.sh names the two kit files it stages, and the kit's scripts source each other

Eleven fixtures under `tests/` build an adopted repository out of this
repository's own kit. Nine copy `.writrun/scripts`, `.writrun/skills`
and `.writrun/templates` as whole trees. Two name the files they stage:
`doctor_lib.sh` named four, and `list_lib.sh` names two —

```
LISTER=".writrun/skills/writrun-select-next-task/list_tasks.sh"
# The lister sources the queue reader from the scripts folder, so the
# fixture carries both — the kit's own layout, not this file's choice.
QUEUE_LIB=".writrun/scripts/stage-2-pull-requests/queue_lib.sh"
```

That comment is the repair after the first time this cost something: the
lister grew a `source` of `queue_lib.sh`, the fixture staged a script
that could not run, and the second line was added by hand. The list is
still a list, so the next dependency the lister grows costs the same
again.

The evidence that it does is on this branch. `doctor_lib.sh` carried the
same shape and never got that repair; WritRun v0.0.09 made
`read_setting.sh` and `check_settings.sh` source `queue_lib.sh`, and
thirty-nine integration cases across `doctor` and `cli` failed with
`queue_lib.sh: No such file or directory`, taking the `release` e2e case
with them. The bump fixed `doctor_lib.sh` by copying the trees, which is
what the kit skill says to do — discover the files, never enumerate
them. `list_lib.sh` was not touched, because it is green and a bump is
not where unrelated fixtures get rewritten.

What is observed is that one fixture is still one `source` away from the
same failure, and that nothing checks for the shape: a fixture naming
kit files passes every gate until a tag disagrees with it.

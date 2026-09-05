---
id: report-0004
status: open
task_ref: []
doc_ref: technical/architecture.md
created: 2026-09-05T13:32:45Z
triaged: null
---

# The fetch is a hybrid the filesystem port does not isolate

**References:** [technical/architecture.md](../../docs/technical/architecture.md)

`kitfetch.Fetch` makes the checkout's directory through the filesystem
port and then hands its path to `git clone`, which fills it on the real
disk. The two halves of one act reach two different filesystems.

Driving `writrun init` or `writrun update` end to end against
`vfs.NewFake()` therefore fails: the fake holds `/tmp/writrun-kit-1` and
the clone writes to a real `/tmp/writrun-kit-1`, so the second run in a
suite hits `fatal: destination path '/tmp/writrun-kit-1' already exists
and is not an empty directory`. Two tests written for the partial-state
messages of `init` and `update` — the ones naming `git checkout -- .`
and `git clean -fd` — were removed for this reason, and those messages
are among the statements no fixture reaches.

The same seam is why task-0016 met every criterion of `spec-0015`
except its coverage floor, which `#29` amends from 99% to 98%.

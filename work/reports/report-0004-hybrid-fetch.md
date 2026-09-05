---
id: report-0004
status: tracked
task_ref: [task-0017]
doc_ref: technical/architecture.md
created: 2026-09-05T13:32:45Z
triaged: 2026-09-05T13:58:35Z
---

# The fetch is a hybrid the filesystem port does not isolate

**References:** [technical/architecture.md](../../docs/technical/architecture.md) · [task-0017](../tasks/task-0017-give-the-fetch.md)

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
except its coverage floor, which #29 amends from 99% to 98%.

**Naming it "the port does not isolate the fetch" reads the seam from
the wrong side.** `git clone` is a subprocess writing to the disk it was
given; no filesystem port will ever stand between it and what it writes.
What the observation actually shows is that the fetch is a boundary of
its own — it leaves the process, exactly as script execution, `gh` and
the terminal do, and `boundaries.md` asks each of those to sit behind a
small interface with a fake beside it. `kitfetch.Fetch` is that
interface already; it has no fake.

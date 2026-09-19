---
id: report-0052
status: tracked
task_ref: [task-0041]
doc_ref: null
created: 2026-09-19T22:45:24Z
triaged: 2026-09-19T23:36:06Z
---

# A signal case can kill the test binary, so an unrelated change goes red

**References:** [task-0041](../tasks/task-0041-signal-keeper.md)

`internal/command/finishcmd/signals_test.go` proves the guard by
sending real signals to the test binary: `raise()` calls
`self.Signal(sig)` on `os.Getpid()`. Its own comment states the
premise — "every case that calls it has something registered for that
signal first ... so the test binary is never killed by its own
fixture". On CI the binary was killed by it.

The unit tier of pull request #174, whose change is `doctorcmd`, the
`doctor` document and the drawings:

```
ok      github.com/thomasfranke/writrun-cli/internal/command/doctorcmd  1.387s
-test.shuffle 1789857791967902218
signal: terminated
FAIL    github.com/thomasfranke/writrun-cli/internal/command/finishcmd  0.554s
make: *** [Makefile:43: cover] Error 1
```

Run
[35474055199](https://github.com/thomasfranke/writrun-cli/actions/runs/35474055199).
0.554s is not the 10-minute timeout `scripts/coverage.sh` sets: the
process was terminated, and SIGTERM is the signal these cases raise.

**The order is not the trigger.** Eight local runs at the seed the
failing job printed — `go test ./internal/command/finishcmd/ -race
-shuffle=1789857791967902218 -count=1` — all pass. What changes between
them is timing, not sequence.

Signal delivery to a Go process is asynchronous, and `signal.Stop` on
the last channel registered for a signal puts that signal back to its
default disposition, which for SIGTERM is death. Several cases raise
and then stop a probe or let the guard's own registration end. That is
the shape the evidence fits; nothing here proves it is the one that
fired.

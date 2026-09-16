# Boundaries

- **Boundaries are interfaces.** Everything that leaves the process —
  script execution, `gh`, the filesystem, the terminal, and this binary
  run again as a process of its own — sits behind a small interface
  defined where it is consumed, with a fake beside it. Production wiring
  happens in `cmd/writrun/` only.
- **The binary is on that list for the same reason as the rest.** A
  command chosen on a screen that asks a question is spawned rather than
  called, so that nothing of it is left reading this process's terminal
  afterwards (decision
  [0015](../decisions/runtime/0015-a-command-that-asks-is-a-process.md)).
  It is a port like any other: a suite drives the routing with a fake
  and never spawns anything.
- **Clean architecture and DDD, applied to size.** Each command package
  is a use case depending inward on ports and nothing else. There is no
  domain layer beyond the queue's vocabulary — task, spec, report,
  stage — because the domain lives upstream in WritRun
  (decision
  [0007](../decisions/engineering/0007-google-style-ports-three-tiers.md)).
- **No global state.** Configuration enters through parameters; nothing
  reads a package-level variable another package wrote.

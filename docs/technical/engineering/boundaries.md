# Boundaries

- **Boundaries are interfaces.** Everything that leaves the process —
  script execution, `gh`, the filesystem, the terminal — sits behind a
  small interface defined where it is consumed, with a fake beside it.
  Production wiring happens in `cmd/writrun/` only.
- **Clean architecture and DDD, applied to size.** Each command package
  is a use case depending inward on ports and nothing else. There is no
  domain layer beyond the queue's vocabulary — task, spec, report,
  stage — because the domain lives upstream in WritRun
  (decision
  [0007](../decisions/engineering/0007-google-style-ports-three-tiers.md)).
- **No global state.** Configuration enters through parameters; nothing
  reads a package-level variable another package wrote.

# `writrun finish`

Ends the work: the promised doc changes are verified, the outcome
recorded, and the pull request marked ready for review.

- [The shape](shape.md) holds: checks first, composition shown, nothing
  on the forge without confirmation.
- What it writes is the spec's `implemented` and the task's `completed`
  date.
- The writes precede the last of its checks, which warns when the
  `completed` date is absent.

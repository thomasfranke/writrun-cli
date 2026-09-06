# `writrun amend`

Reopens an approved spec that must change: it returns to draft for
re-approval, and the pull request says why.

- [The shape](shape.md) holds: checks first, composition shown, nothing
  on the forge without confirmation.
- What it writes is `status: draft` on the named spec, and nothing on
  the task.
- The pull request opens ready for review — an amendment announces no
  work.
- The body names the pull request the amendment suspends.

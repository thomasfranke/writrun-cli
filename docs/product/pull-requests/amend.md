# `writrun amend`

Reopens an approved spec that must change: it returns to draft for
re-approval, and the pull request says why.

- [The shape](shape.md) holds.
- The composed title and body are handed to the repository's own
  observance check before the branch is cut, so a title outside the
  declared vocabulary is refused here rather than on the forge.
- What it writes is `status: draft` on the named spec, and nothing on
  the task.
- The pull request opens ready for review — an amendment announces no
  work.
- The body names the pull request the amendment suspends.

# `writrun finish`

Ends the work: the promised doc changes are verified, the outcome
recorded, and the pull request marked ready for review.

- [The shape](shape.md) holds: checks first, composition shown, nothing
  on the forge without confirmation.
- What it writes is the spec's `implemented` and the task's `completed`
  date.
- The writes precede the last of its checks, which warns when the
  `completed` date is absent.
- **Its gates judge the tree they vouch for.** The writes land in the
  working tree and the command commits nothing, so a gate reading the
  branch as last committed passes having seen none of them — and what
  it vouched for is the branch as it was before the command ran. The
  range `finish` hands its checks reaches the working tree, so the spec
  it has just set `implemented` and the task it has just dated are
  inside what is judged.
- A gate that cannot see the edits is worse than a gate that refuses
  them: a refusal is read and answered, and a pass over an unread change
  is a sentence nobody knows is empty.

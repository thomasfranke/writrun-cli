# `writrun finish`

Ends the work: the promised doc changes are verified, the outcome
recorded, and the pull request marked ready for review.

- [The shape](shape.md) holds: checks first, composition shown, nothing
  on the forge without confirmation.
- What it writes is the spec's `implemented` and the task's `completed`
  date.
- The writes precede the last of its checks, which warns when the
  `completed` date is absent.
- **Its gates judge the tree they vouch for**, as far as they can be
  made to. The writes land in the working tree and the command commits
  nothing, so a gate reading the branch as last committed passes having
  seen none of them — and what it vouched for is the branch as it was
  before the command ran.
- **The first gate reads the tree; the last two do not yet.** The delta
  check is called directly and is handed a bare range, whose diff
  reaches the working tree — so a promised document written and staged
  is judged rather than refused. The gates after the writes run through
  `preflight.sh`, which takes an argument holding `..` as its range and
  anything else as a task list; the bare shape is the only one that
  reaches the tree, and it is the one shape that script will not accept.
  Until that changes upstream, those two still read the branch as
  committed and still answer `deltas checked: none`
  (`work/reports/report-0046-preflight-bare-range.md`).
- A gate that cannot see the edits is worse than a gate that refuses
  them: a refusal is read and answered, and a pass over an unread change
  is a sentence nobody knows is empty.

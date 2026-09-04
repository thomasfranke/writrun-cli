# `writrun status`

Answers where the work stands, from the current branch.

- Names the task the current branch carries — or that it carries
  none — with its spec and the spec's status.
- Runs the completion checks read-only and names the first that would
  fail — what [`finish`](../pull-requests/finish.md) would stop at.
- Counts the open reports awaiting triage.
- Compares the kit's recorded tag with the tag this client pins; a
  mismatch is named, never bridged.
- Reads only — nothing changes, on the repository or the forge.

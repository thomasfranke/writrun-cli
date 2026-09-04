# `writrun report`

Records an observation into `work/reports/` before it is lost to a
conversation.

- Takes the observation as given — a title and the body the reporter
  wrote — and writes the file through the methodology's own generator:
  the id minted in sequence and never reused, the status born `open`,
  nothing triaged yet.
- Records only. **Triage — choosing what becomes of the report — is a
  judgement**, and judgements belong to a human or an agent, never to
  this command. No flag sets a route.
- Writes the file wherever you are: the methodology exempts recording
  from the one-kind-per-change rule, so no branch is created and no
  pull request is opened.
- A recorded report changes nothing else — the queue, the statuses,
  and the forge are exactly as they were.

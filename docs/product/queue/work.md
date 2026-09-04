# `writrun work`

Selects a task and launches the adopter's agent on it.

- With no argument, takes the next available task by the algorithm.
  With a task named, uses that one — and refuses if it is not
  available, naming why.
- Launches **the agent command the adopter configured** — `writrun`
  never guesses which agent is installed ([rules](../rules.md)).
- With no agent configured, aborts and shows exactly how to configure
  it. Nothing is launched and nothing changes.
- Points the agent at the project's agent instructions and the selected
  task, so it starts with the brief the methodology already assembled.
- Takes nothing itself: the launched agent opens the draft the
  methodology requires, exactly as a human would with
  [`take`](../pull-requests/take.md).
- The gates are unchanged for whatever it launches: an agent started
  this way approves nothing.

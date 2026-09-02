# Queue — `list`, `work`

## `writrun list`

Reads the queue without typing script paths.

- Shows two groups: what is **available** to take now, and what is
  **held back** — each held-back task with the reason it is held.
- The order of the available group is the methodology's selection
  algorithm, unchanged.
- Reads the queue files as the authority. Any forge mirror is a
  projection, and its absence never changes the answer.
- Filters narrow the listing; they never change how eligibility is
  decided.
- Reads only — nothing about the queue changes.

## `writrun work`

Selects a task and launches the adopter's agent on it.

- With no argument, takes the next available task by the algorithm.
  With a task named, uses that one — and refuses if it is not
  available, naming why.
- Launches **the agent command the adopter configured**. `writrun`
  never guesses which agent is installed, and never acts as one.
- With no agent configured, aborts and shows exactly how to configure
  it. Nothing is launched and nothing changes.
- Points the agent at the project's agent instructions and the selected
  task, so it starts with the brief the methodology already assembled.
- The gates are unchanged for whatever it launches: an agent started
  this way approves nothing.

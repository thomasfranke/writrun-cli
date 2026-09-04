---
id: spec-0007
task_ref: task-0007
status: approved
created: 2026-09-03T22:30:39Z
---

# spec-0007 — Select the task and launch the configured agent

**References:** [task-0007](../tasks/task-0007-work-command.md)

- **Goal:** `writrun work` selects and launches; it never reasons about the task.

## Scope

In: selection (next available, or the named task), brief assembly,
launching the configured agent. Out: acting as an agent; any queue
write; any approval.

## Steps

1. Read the agent command from `git config writrun.agent`; absent — abort printing the exact `git config` line to set it, launching nothing.
2. No argument: first task of the selection algorithm's available group. Named: verify availability; refuse an unavailable task with the lister's own held-back reason.
3. Assemble the brief with `brief.sh <task-id>`.
4. Launch the configured command pointed at `AGENTS.md` and the brief; gates are unchanged for whatever is launched.

## Acceptance criteria (EARS)

- When no agent is configured, the system shall abort showing how to configure one, and launch nothing.
- When the named task is not available, the system shall refuse naming the reason.
- When nothing is available, the system shall say so and launch nothing.
- When launched, the agent shall receive the brief `brief.sh` assembled, unedited.

## Edge cases

- `brief.sh` exit 2 (partial brief): shown, and the launch is refused — an incomplete brief is not a working brief.
- Agent command exits non-zero: passed through as the command's own exit.

## Tests required

Integration with a stub agent command recording its invocation.

## Definition of Done

- [ ] A stub agent receives task id, brief and instructions for every fixture.
- [ ] Suite green.

## Proposed product changes

- none — `product/queue/work.md` already states the behaviour.

## Proposed technical changes

- none — `technical/architecture.md` already states the config key.

## Outcome

_(fill after execution)_

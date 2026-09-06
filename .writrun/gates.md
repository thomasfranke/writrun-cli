# Human gates — per principle 7

This file is the project's, like `settings.json`: `writ update` never
touches it, and every project states its own answers here. Fill the
TODOs — naming an agent as the operator of a gate is a valid answer;
leaving a gate unnamed is not.

The routing row is this project's own, added to the seven the kit
states: it is the yes `.writrun/AGENTS.md` requires before a report
becomes an issue upstream ([`AGENTS.md`](../AGENTS.md)).

| Transition | Who |
|---|---|
| Writing or changing anything under `docs/` | Thomas writes or reviews before merge. |
| An authored rule is finished, so derivation may start | Thomas declares it. |
| Spec `draft → approved` | Thomas only, recorded via the approved PR. |
| Task with empty `spec_ref` and insufficient brief | Stop and ask for a spec. |
| Derived work, before the PR opens | Present it in the session before the PR opens. |
| Changing repository/forge settings (Actions permissions, rulesets, merge methods) | Thomas assents in session, per set of changes. |
| A report becomes a task (`tracked`) | The agent derives; the merge of that `report/` branch assents. |
| A report is routed to the WritRun repository | Thomas says yes, per report, before the issue opens. |
| Everything else | Agent, autonomously. |

**The forge row is not optional the way its answer is.** Repository
settings live outside the repository — no diff, no review, no merge gate
sees them — so an agent applying one is acting where nothing can catch
it afterwards. Whoever the project names, the agent presents current →
target values first and applies only on an explicit yes.

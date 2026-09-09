# Human gates — per principle 7

**This is WritRun's default answer, not your project's.** It is in force
while `writrun/gates.md` defers to it, and every gate below is answered
the cautious way: a human operates it. Replace that file's content to
answer differently — naming an agent as the operator of a gate is a
valid answer, and the file you write is yours whole, never merged with
this one.

| Transition | Who |
|---|---|
| Writing or changing anything under `docs/` | Human writes, or reviews before merge; agents may draft. |
| An authored rule is finished, so derivation may start | **Human declares it** — never inferred. |
| Spec `draft → approved` | **Human only**, recorded through the merge of the change that proposes it. |
| Task with empty `spec_ref` and insufficient brief | **Stop and ask for a spec** — never improvise scope. |
| Derived work, before the PR opens | **Present it in the session.** |
| Changing repository/forge settings (Actions permissions, rulesets, merge methods) | **Owner assents in session**, per set of changes. |
| A report becomes a task (`tracked`) | **Agent derives, human assents** — through that change's own merge. |
| Everything else | Agent, autonomously. |

**Why the defaults sit here rather than in your file.** A gate nobody
answered used to be a TODO in the adopter's home, which no update could
correct and no reader was obliged to fill; the cost of never opening
the file was a project whose gates were, by its own terms, unanswered.
Deferring costs the opposite and cheaper mistake — asking a human more
often than that project needs — and the fix is one file the project
writes when it wants it
([gates](https://github.com/thomasfranke/writrun/blob/main/docs/product/stage-1-tasks-and-specs/gates.md#what-a-projects-own-table-has-to-carry)).

**The forge row is not optional the way its answer is.** Repository
settings live outside the repository — no diff, no review, no merge gate
sees them — so an agent applying one is acting where nothing can catch
it afterwards. Whoever the project names, the agent presents current →
target values first and applies only on an explicit yes.

---
id: report-0026
status: fixed
task_ref: []
doc_ref: product/pull-requests/shape.md
created: 2026-09-06T04:25:03Z
triaged: 2026-09-06T07:36:44Z
---

# shape.md lists an order that neither finish nor amend literally follows

**References:** [product/pull-requests/shape.md](../../docs/product/pull-requests/shape.md)

`product/pull-requests/shape.md` opens by giving `author`, `take`,
`finish` and `amend` one shape: "run the repository's checks in the
order the methodology fixed, write the status the flow calls for,
assemble the branch, the commit title, and the pull-request body, show
all of it, and open the pull request on confirmation." Two of the four
do not follow that order literally. `finish` writes at step 2 and runs
`preflight.sh` after — the timing report-0019 corrected on
`finish.md` — and `amend` computes its queue edit at step 2 but writes
it beside the push, after the confirmation, which `amendcmd`'s package
comment states and which spec-0011's Outcome records as required rather
than optional. The sentence survives on its own "in the order the
methodology fixed" deferral, which declines to fix the internal order
and so cannot be contradicted by one. Whether a sentence that four
commands are measured against should be that elastic is what this
records; report-0019 repaired the page that carried no such deferral,
and left this one standing.

**Triage — fixed.** The sentence was replaced, not repaired. No order
the four commands share exists, so no ordered wording could have been
made true of all of them.

Four of its five clauses were false of at least one command. `amend`
ran no repository check at all before task-0024; `take` and `author`
write no status, and `take_task.sh` says so itself; `take`, `author`
and `finish` compose no commit title; `finish` assembles no branch and
opens no pull request. Only "show all of it" held for all four.

The deferral does not carry the rest of the series. "In the order the
methodology fixed" attaches to the checks — that is the clause it
modifies in `product/rules.md`, where the fact lives — so the four
clauses after it were unqualified claims. spec-0017's Outcome had
meanwhile cited the sentence as a binding inter-clause order, which is
the wide reading the code never obeyed.

What replaces it says what the page is for: the four commands end a
flow in a pull request, the bullets below hold for all four, and what
each one checks, writes and opens is on its own page. The old sentence
also failed the style — five claims in one sentence, and an
enumeration of five items written as a paragraph.

Bullet 1's heading went with it. "Checks first, in their load-bearing
order" carried the same falsehood in miniature: `finish` writes before
the last of its checks, and task-0024 gives `amend` and `author` a
check whose input is the composition. It is now two bullets — a
non-zero check stops the command there, and a check runs where its
input exists.

Bullet 4's "Nothing reaches the forge without confirmation" is loose
the same way and is left standing: `amend` calls `gh pr list` and
`finish` calls `gh pr view` before their questions. It is not this
change's, and it is recorded nowhere yet.

`take.md` and `finish.md` still recite "checks first, composition
shown, nothing on the forge without confirmation", which is now a
phrase `shape.md` does not carry. `amend.md` and `author.md` dropped
the recital because task-0024's spec named them; the other two pages
were outside it.

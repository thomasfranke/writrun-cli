---
id: report-0044
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-17T11:37:06Z
triaged: 2026-09-17T11:37:24Z
---

# two meanings shared one field until doctor became a screen

`Command.AsksNothing` answered two questions that were the same answer
for as long as nobody noticed:

- may the screen capture this command's output and page it, or must the
  terminal be handed over;
- is this command run so often that explaining it on a bare run would be
  noise.

Four commands answered yes to both — `list`, `status`, `doctor`,
`work` — so one field served, and spec-0039's implementation said so
where it stood: *"AsksNothing is that same set, already declared."*

task-0033 and task-0035 were worked at the same time, on branches that
never saw each other. task-0033 gave `doctor` a screen, which means
`doctor` reads the terminal — so it stopped declaring `AsksNothing`,
correctly, by that field's own documented meaning. task-0035 gated the
bare-run explanation on the same field. Neither branch was wrong and
neither could see the other; the two met in the rebase.

The result would not have been noise. A bare `writrun doctor` would
print the whole explanation into the normal buffer and then enter the
alternate one, which wipes it — so the reader sees nothing, does their
work, leaves the screen, and meets twenty lines of introduction on the
way out, for a command they run every day.

Nothing would have said so. `git` reported one conflict, in the
literal lines both branches edited, and resolving that conflict leaves
this one standing.

**Triage — fixed.** `Daily` is a field of its own, declared by the
four commands run bare, and the bare-run gate reads it.
`AsksNothing` keeps the one meaning its documentation always gave it.
`TestTheDailyCommandsAreTheOnesRunBare` holds the set by name over the
production table — it fails saying `doctor is run bare and does not
declare Daily` if that declaration is ever dropped again.

What outlives the fix is how it was found. Two branches, two approved
specs, a clean textual merge, and the defect lives in neither diff — it
lives in the sentence one of them wrote about a field the other
redefined. No gate here reads a comment.

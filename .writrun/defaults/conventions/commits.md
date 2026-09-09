# Commits

**The commit subject is a constant** — [Conventional
Commits](https://www.conventionalcommits.org/), `type(scope): imperative
summary` — in every project, whatever
[`settings.json`](../../../writrun/settings.json)'s `stage_2.pr_title_style` declares.
That key reaches the pull request title and stops there: a queue of open
pull requests is read by the people working it, while `main` is read by
bisect, by release tooling and by whoever arrives in a year, and that
audience is the same everywhere
([0063](https://github.com/thomasfranke/writrun/blob/main/docs/technical/decisions/pull-requests/0063-title-and-subject-are-two-texts.md)).
This file explains the two vocabularies the subject uses; what a title
does with the same two is [prs.md](prs.md)'s.

- **Types** — the kind of change: `stage_2.commit_types` in
  [`settings.json`](../../../writrun/settings.json).
- **Scopes** — the subsystem it is about, optional: omit when a change
  genuinely spans the repository. `stage_2.commit_scopes` in the same
  file.
- Example: `docs(product): add the coverage-rule concept chapter`.

**The words themselves live in the settings file, and only there.** A
project's vocabulary is a value, so it sits where the machinery and the
reader find one statement — `check_observance.sh` reads exactly the
list the project declared, and a convention that spelled the words a
second time would be a copy free to disagree with the door
([schema](https://github.com/thomasfranke/writrun/blob/main/docs/technical/settings/schema.md#settings)).
Adding a scope is editing one line of `settings.json`, and every check
sees it at once. Read the two lists in force with:

```bash
bash .writrun/scripts/stage-2-pull-requests/read_setting.sh stage_2.commit_scopes
```

**Both lists are lower case, and they are not interchangeable.** A
subject spells them exactly as declared; a bracketed title carries the
same words in whatever case it likes, and `writrun check` folds it.
What neither may do is swap the slots — a scope leading a title is a
type the vocabulary does not have, and the door says so.

**The `[TASK-NNNN]` tags lead the title and stay out of the subject.**
On the title they sit **outside** whichever grammar follows, so the one
parser that exists never has to know both at once: from Stage 2,
`writrun check` strips the tags and reads what is left against the
declared style — the type against `commit_types`, the scope against
`commit_scopes` when one is present, and nothing at all about the
summary that follows. What lands on `main` carries no tag; the `(#NN)` the forge
appends to a squash subject is the hop back to the pull request, which
still carries them and is still what the machinery parses. Anything
downstream of that is read by eye rather than by a strict parser; the
release notes the forge generates come from pull requests and parse
nothing here.

- Trivial work is a commit, never a task (principle 6).

**Three workflows commit, and their subjects are the constant above.**
`writrun approve` records what a merge decided — the specs it approved,
and the `queued`/`merged` dates it earned — `writrun progress` records
what a pull request event moved, and `writrun intake` records the
report a maintainer's label let in. All three take their subject from
[`commit_subject.sh`](../../scripts/stage-2-pull-requests/commit_subject.sh),
one literal per event under the scope `queue`; the text lives there and
not in any workflow, because three callers writing it separately are
three places to edit, and nothing squashes these — a subject that
drifted would sit on `main` for good. Each is one commit because each
records one event.

**A branch's own subjects are a convention kept by hand.** Squash-only
discards every one of them — what reaches `main` is a single subject,
seeded by the pull request title and typed in the merge box — so
`writrun check` reads that title and nothing else. A subject on a branch
is a courtesy to whoever reads the branch, and a gate that failed a pull
request over text the merge discards would be enforcing where nothing is
left behind. **What a branch loses is the door, never the rule**: kept by
hand is still kept, and the shape it is kept in is the constant above.

## Who presses commit — `stage_2.auto_commit`

`true`, the default, lets an agent commit on its own as its flow requires.
`false` gates the action and never the work: the agent still composes the
**whole** message — subject, body, trailers — presents it, and commits only
after an explicit yes. Approval is per commit; several in one working
session are several asks, never one session-wide grant.

**The flag outranks the agent platform's own autonomy mode.** An agent
running auto-accept, autonomous, or any mode in which its harness would not
ask, still stops here: the platform's mode governs what the *harness* asks,
this flag governs what the *adopter* allowed. A setting that only bound an
agent already asking would control nothing.

Neither this flag nor `auto_pr` touches the machinery's own commit above,
nor any workflow-driven write — those are not an agent's actions.

## Whether the agent signs — `stage_2.agent_coauthor`

`true`, the default, obliges an agent to append a `Co-Authored-By:`
trailer **naming the model** to every commit it writes:

```
Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>
```

`false` means the message carries the change alone — no co-author
trailer, no session URL, no tool mention; it reads as any other in the
history. An instruction from the agent's own platform to append credit
yields to this file, either way, with the same precedence `auto_commit`
states.

**`true` states a shape, not a permission**, and that is what changed
when the key stopped being called `credit_ai`. The old wording left the
agent whatever credit its *platform* appended — a source, not an
artifact, so there was nothing for a check to look for and nothing an
agent on a silent platform owed. Naming the trailer makes the obligation
the agent's: on a platform that appends no credit of its own, the agent
**writes** the trailer rather than having nothing to keep.

**The model is named specifically, not as a category** — `Claude Opus 5`,
never `AI` or `an agent`. A record that survives the next model's arrival
is the whole reason the trailer is worth reading a quarter later, and a
category name answers nothing.

The flag speaks only to what an agent writes. Authorship and committer
identity stay git configuration, nobody else's commits are touched, and
nothing rewrites history — it binds from the commit after the flip.

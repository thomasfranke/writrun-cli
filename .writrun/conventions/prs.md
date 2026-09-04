# Pull requests

- **Title**: every task the PR carries, tagged, then a summary written in
  the style [`settings.json`](../settings.json) names.

  **The tag is not the settable part.** One bracket per task, uppercase,
  no separator — `[TASK-0012]`, or `[TASK-0012][TASK-0014]` when several.
  It leads because a title is read in a list of open pull requests the
  way a subject is read in a log — down the left edge — so what work this
  was comes before what kind of change it was. It is also how the
  machinery learns which tasks the PR carries, so it stays whatever the
  style. Authoring and reporting PRs carry none — their tasks are born in
  the PR, not worked by it.

  **`stage_2.pr_title_style` chooses what follows**, and the choice is
  about the people who read the queue — the open pull requests, which are
  worked by the team that wrote them:

  - `conventional` — the same grammar the commits use:
    `[TASK-0012] fix(ci): debounce mirror updates`, and
    `docs(product): the merge is the assenting act` for an authoring PR.
    One shape to learn, and a queue that scans like the log beside it.
  - `bracketed` — a human sentence behind bracketed labels:
    `[TASK-0012][Fix][CI] Debounce mirror updates`, and
    `[DOCS] The merge is the assenting act`. The labels are read against
    the same two vocabularies, and the sentence after them is not parsed
    at all — so this costs no guarantee; it reads as prose rather than as
    a grammar, which suits a queue read mostly by people.

  **What lands on `main` is neither of these, and is not settable.** The
  squash dialog's subject is the merging maintainer's to type — the title
  only seeds it — and the grammar there is Conventional Commits in every
  project, [`commits.md`](commits.md)'s constant. The `[TASK-NNNN]` tag
  goes no further than the title; the `(#NN)` the forge appends to the
  subject is the hop back to the pull request that carries it
  ([0063](https://github.com/thomasfranke/writrun/blob/main/docs/technical/decisions/pull-requests/0063-title-and-subject-are-two-texts.md)).

  **Whichever is declared, `writrun check` reads the title against it**
  from Stage 2 on — the style is a setting an agent was told to obey, and
  a title is where disobeying it leaves a trace
  ([observance](https://github.com/thomasfranke/writrun/blob/main/docs/technical/README.md#observance-is-checked-where-it-leaves-a-trace)).
  Case inside the brackets is not judged: `[Fix]` and `[DOCS]` are both
  the vocabulary, spelled two ways this file itself uses.

  Neither is more correct. Pick the one your readers already know, state
  it in the settings file, and let the agents follow it.
- **Body**: the [template](../templates/pull_request_template.md), lives
  only in `.writrun/templates/` — agents fill it when opening any PR; a
  human opening one by hand copies it from there (GitHub does not
  pre-fill from `.writrun/`; this project chose one home over the
  platform's pre-fill). Everything in it is editable except the
  `## Derived work` heading, which `writrun check` reads — a **contract
  marker**.
- **Opening state**: an implementing PR opens as a **draft**, at the
  moment its task is taken and before the work starts — that is what puts
  the task's mirror on `status:in-progress`. Ready for review is the end
  of the work, not the start. Authoring and reporting PRs have no work to
  announce and open ready.
- **The push and the opening are one act.** The branch reaches the forge
  and the draft opens in the same breath, so what an adopter gates is the
  moment their work becomes public — not the second half of it. The agent
  composes the branch name, the title and the body, presents them
  together, and does both only on the word. See
  [`auto_push`](#who-pushes--stage_2auto_push) below.
- **Merge**: squash only — a messy branch history is fine; the commit
  landing on `main` is not.

## Who pushes — `stage_2.auto_push`

`true`, the default, lets an agent push on its own as its flow requires.
`false` gates the action and never the work: the agent commits as it
would, names the branch, and pushes only after an explicit yes.

**This is the flag that holds work in.** A commit is private and a pull
request is already a conversation; the push between them is the act that
puts an adopter's work on someone else's server, and it is the last
moment anything can still be taken back quietly. Before this key existed
it belonged to no flag at all — a first push was nobody's, and a push to
an open pull request's branch read as `auto_pr`'s by inference.

**Before the pull request exists, this flag and `auto_pr` are one
question.** Taking a task pushes and opens the draft in one act, so the
agent asks once, presenting the branch, the title and the body together;
`false` on either flag holds the whole thing. Two prompts for one moment
is not a stricter gate. Once the pull request is open, further pushes to
its head branch are this flag's alone.

**It reaches every branch, not only a task's.** A docs branch, a report
branch and an implementing branch make work public the same way, and the
gate is about that, never about what kind of work it is.

## Who opens the pull request — `stage_2.auto_pr`

`true`, the default, lets an agent open and update pull requests on its
own as its flow requires. `false` gates the action and never the work: the
agent still composes the **complete** title and body — the template
filled, the specs named, the verification stated — presents them, and
opens the pull request only after an explicit yes. Approval is per action,
never a session-wide grant.

**The flag outranks the agent platform's own autonomy mode**, exactly as
`auto_commit` does in [commits.md](commits.md): the platform's mode
governs what the *harness* asks, this flag governs what the *adopter*
allowed.

**It holds the draft too.** The pull request that reports a task as taken
is mechanical, but mechanical is not exempt — the flag gates the action,
not the reason for it. The agent presents, waits, and the taking flow
continues unchanged afterwards. It holds the push that carries the branch
there with it: at that moment the two are one act
([`auto_push`](#who-pushes--stage_2auto_push)), and gating only the
second half would leave the branch on the forge before the adopter was
asked anything. It sits in `stage_2`, beside
`auto_commit`, because git begins at Stage 2; below that there is
neither a commit nor a pull request for either flag to gate.

`stage_2.agent_coauthor` reaches the body as it reaches a commit
message. With `false`, nothing an agent writes into the forge carries a
generated-with line, a session URL or any other credit — and from Stage 2
that is checked, over the pull request's own commits and body. With
`true`, the body carries a credit line and the commits carry the
model-naming trailer [commits.md](commits.md) describes; **the commits
are what a check reads**, because a trailer has a fixed shape and a fixed
place and a body's credit line has neither, so the body's half stays
instruction-bound and is stated as such rather than guessed at.

What no diff can show — whether the agent *asked* before committing,
pushing or opening — stays instruction-bound too, and no check pretends
otherwise.
